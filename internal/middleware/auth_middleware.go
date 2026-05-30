package middleware

import (
	"net/http"
	"os"
	"strings"

	"cerdasind-backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Authorization header required"})
			c.Abort()
			return
		}

		// Handle both "Bearer <token>" and raw "<token>"
		tokenString := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			tokenString = strings.TrimSpace(authHeader[7:])
		} else {
			tokenString = strings.TrimSpace(authHeader)
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Invalid authorization format"})
			c.Abort()
			return
		}

		secret := os.Getenv("JWT_SECRET")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Invalid or expired token"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", int64(claims["user_id"].(float64)))
		c.Set("role", model.UserRole(claims["role"].(string)))
		c.Next()
	}
}

func RoleMiddleware(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		userRole := role.(model.UserRole)

		allowed := false
		for _, r := range roles {
			if userRole == r {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, model.ErrorResponse{Message: "Access forbidden: insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}
