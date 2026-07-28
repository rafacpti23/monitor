package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"p-mon-backend/internal/config"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int64, role string) (string, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour) // 30 days
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Cfg.JWTSecret))
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Cfg.JWTSecret), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, fmt.Errorf("invalid signature")
		}
		return nil, fmt.Errorf("invalid token: %v", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			tokenString = c.Query("token") // Check query param for WebSockets/downloads
		} else {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func GetUserID(c *gin.Context) int64 {
	id, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return id.(int64)
}

func GetUserRole(c *gin.Context) string {
	role, exists := c.Get("role")
	if !exists {
		return ""
	}
	return role.(string)
}
