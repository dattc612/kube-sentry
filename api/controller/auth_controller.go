package controller

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/dattc/go-though-storm/api/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthController interface {
	LoginHandler(ctx *gin.Context)
}

type authController struct {
	auth   models.LocalAuthConfig
	logger *slog.Logger
}

func NewAuthController(auth models.LocalAuthConfig, logger *slog.Logger) AuthController {
	if logger == nil {
		logger = slog.Default()
	}
	return &authController{
		auth:   auth,
		logger: logger,
	}
}

func (a *authController) LoginHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		a.logger.WarnContext(ctx, "invalid login request payload", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	var matchedUser *models.UserCredentials
	for _, u := range a.auth.Users {
		if u.Username == body.Username {
			matchedUser = &u
			break
		}
	}

	if matchedUser == nil || bcrypt.CompareHashAndPassword([]byte(matchedUser.Password), []byte(body.Password)) != nil {
		a.logger.WarnContext(ctx, "failed login attempt", slog.String("username", body.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	expirationTime := time.Now().Add(8 * time.Hour)
	claims := &models.Claims{
		Username: body.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		Roles: matchedUser.Roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.auth.JwtSecret)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to generate JWT token",
			slog.String("username", body.Username),
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	a.logger.InfoContext(ctx, "user logged in successfully", slog.String("username", body.Username))

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
