package controller

import (
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
	auth models.AuthConfig
}

func NewAuthController(auth models.AuthConfig) AuthController {
	return &authController{
		auth: auth,
	}
}

func (a *authController) LoginHandler(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	expirationTime := time.Now().Add(8 * time.Hour)
	claims := &models.Claims{
		Username: body.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.auth.JwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
