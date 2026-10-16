package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/dattc/go-though-storm/api/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	Sub               string   `json:"sub"`
	Email             string   `json:"email"`
	PreferredUsername string   `json:"preferred_username"`
	RealmRoles        []string `json:"-"`
	ClientRoles       []string `json:"-"`
}

type keycloakClaims struct {
	jwt.RegisteredClaims
	Email             string         `json:"email"`
	PreferredUsername string         `json:"preferred_username"`
	RealmAccess       realmAccess    `json:"realm_access"`
	ResourceAccess    resourceAccess `json:"resource_access"`
}

type realmAccess struct {
	Roles []string `json:"roles"`
}

type resourceAccess map[string]struct {
	Roles []string `json:"roles"`
}

type contextKey string

const userContextKey contextKey = "user"

func GetUser(c *gin.Context) (*TokenClaims, bool) {
	user, exists := c.Get(string(userContextKey))
	if !exists {
		return nil, false
	}
	claims, ok := user.(*TokenClaims)
	return claims, ok
}

func keycloakAuthMiddleware(client *gocloak.GoCloak, cfg *models.KeycloakAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing or malformed authorization header",
			})
			return
		}

		ctx := c.Request.Context()

		jwtToken, claimsMap, err := client.DecodeAccessToken(ctx, tokenStr, cfg.Realm)
		if err != nil || jwtToken == nil || !jwtToken.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired access token",
			})
			return
		}

		claims, err := parseTokenClaims(tokenStr, cfg.ClientID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": fmt.Sprintf("Failed to parse token claims: %v", err),
			})
			return
		}

		// 1. Extract Identity (Preferred Username -> Subject ID)
		var identity string
		if claimsMap != nil {
			if preferredUsername, ok := (*claimsMap)["preferred_username"].(string); ok && preferredUsername != "" {
				identity = preferredUsername
			} else if sub, ok := (*claimsMap)["sub"].(string); ok && sub != "" {
				identity = sub
			}
		}

		// 2. Extract Roles (Realm Roles & Client Roles)
		var roles []string
		if claimsMap != nil {
			// Realm roles
			if realmAccess, ok := (*claimsMap)["realm_access"].(map[string]interface{}); ok {
				if rList, ok := realmAccess["roles"].([]interface{}); ok {
					for _, r := range rList {
						if roleStr, ok := r.(string); ok {
							roles = append(roles, roleStr)
						}
					}
				}
			}

			// Client-specific roles
			if resourceAccess, ok := (*claimsMap)["resource_access"].(map[string]interface{}); ok {
				if clientAccess, ok := resourceAccess[cfg.ClientID].(map[string]interface{}); ok {
					if rList, ok := clientAccess["roles"].([]interface{}); ok {
						for _, r := range rList {
							if roleStr, ok := r.(string); ok {
								roles = append(roles, roleStr)
							}
						}
					}
				}
			}
		}

		c.Set(string(userContextKey), claims)
		c.Set("user_id", identity)
		c.Set("username", identity)
		c.Set("roles", roles)

		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetUser(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Not authenticated",
			})
			return
		}

		allRoles := make(map[string]bool)
		for _, r := range claims.RealmRoles {
			allRoles[r] = true
		}
		for _, r := range claims.ClientRoles {
			allRoles[r] = true
		}

		for _, required := range roles {
			if !allRoles[required] {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": fmt.Sprintf("Missing required role: %s", required),
				})
				return
			}
		}

		c.Next()
	}
}

func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetUser(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Not authenticated",
			})
			return
		}

		allRoles := make(map[string]bool)
		for _, r := range claims.RealmRoles {
			allRoles[r] = true
		}
		for _, r := range claims.ClientRoles {
			allRoles[r] = true
		}

		for _, allowed := range roles {
			if allRoles[allowed] {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
	}
}

func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("authorization header must be Bearer {token}")
	}
	return parts[1], nil
}

func parseTokenClaims(tokenStr string, clientID string) (*TokenClaims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenStr, &keycloakClaims{})
	if err != nil {
		return nil, err
	}

	kc, ok := token.Claims.(*keycloakClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type")
	}

	var clientRoles []string
	if access, exists := kc.ResourceAccess[clientID]; exists {
		clientRoles = access.Roles
	}

	return &TokenClaims{
		Sub:               kc.Subject,
		Email:             kc.Email,
		PreferredUsername: kc.PreferredUsername,
		RealmRoles:        kc.RealmAccess.Roles,
		ClientRoles:       clientRoles,
	}, nil
}

func localAuthMiddleware(cfg models.LocalAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing or malformed authorization header",
			})
			return
		}

		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set(string(userContextKey), claims)
		c.Set("username", claims.Username)
		c.Set("user_id", claims.Username)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

func dualAuthMiddleware(keycloakClient *gocloak.GoCloak, keycloakCfg *models.KeycloakAuthConfig, localCfg models.LocalAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing or malformed authorization header",
			})
			return
		}

		ctx := c.Request.Context()

		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(localCfg.JwtSecret), nil
		})

		if err == nil && token.Valid {
			c.Set(string(userContextKey), claims)
			c.Set("username", claims.Username)
			c.Set("user_id", claims.Username)
			c.Set("roles", claims.Roles)
			c.Next()
			return
		}

		jwtToken, claimsMap, err := keycloakClient.DecodeAccessToken(ctx, tokenStr, keycloakCfg.Realm)
		if err == nil && jwtToken != nil && jwtToken.Valid {
			kcClaims, parseErr := parseTokenClaims(tokenStr, keycloakCfg.ClientID)
			if parseErr == nil {
				var identity string
				if claimsMap != nil {
					if preferredUsername, ok := (*claimsMap)["preferred_username"].(string); ok && preferredUsername != "" {
						identity = preferredUsername
					} else if sub, ok := (*claimsMap)["sub"].(string); ok && sub != "" {
						identity = sub
					}
				}

				var roles []string
				if claimsMap != nil {
					if realmAccess, ok := (*claimsMap)["realm_access"].(map[string]interface{}); ok {
						if rList, ok := realmAccess["roles"].([]interface{}); ok {
							for _, r := range rList {
								if roleStr, ok := r.(string); ok {
									roles = append(roles, roleStr)
								}
							}
						}
					}

					if resourceAccess, ok := (*claimsMap)["resource_access"].(map[string]interface{}); ok {
						if clientAccess, ok := resourceAccess[keycloakCfg.ClientID].(map[string]interface{}); ok {
							if rList, ok := clientAccess["roles"].([]interface{}); ok {
								for _, r := range rList {
									if roleStr, ok := r.(string); ok {
										roles = append(roles, roleStr)
									}
								}
							}
						}
					}
				}

				c.Set(string(userContextKey), kcClaims)
				c.Set("username", identity)
				c.Set("user_id", identity)
				c.Set("roles", roles)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: Invalid token for both Local and Keycloak authentication",
		})
	}
}

func AuthMiddleware(auth models.AuthConfig, logger *slog.Logger) gin.HandlerFunc {
	logger.Info("configuring authentication middleware", slog.String("mode", auth.Mode))

	switch auth.Mode {
	case "keycloak":
		client := gocloak.NewClient(auth.Keycloak.URL)
		return keycloakAuthMiddleware(client, &auth.Keycloak)
	case "local":
		return localAuthMiddleware(auth.Local)
	case "dual":
		client := gocloak.NewClient(auth.Keycloak.URL)
		return dualAuthMiddleware(client, &auth.Keycloak, auth.Local)
	default:
		logger.Error("unsupported auth mode requested", slog.String("mode", auth.Mode))
		panic("Unsupported auth mode: " + auth.Mode)
	}
}
