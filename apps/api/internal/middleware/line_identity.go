package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const lineUserIDKey = "linora.lineUserID"

type LineIdentityVerifier interface {
	VerifyIDToken(context.Context, string) (string, error)
}

func RequireLineIdentity(verifier LineIdentityVerifier, environment string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.EqualFold(environment, "development") {
			if userID := strings.TrimSpace(c.GetHeader("X-Linora-Dev-User")); userID != "" {
				c.Set(lineUserIDKey, userID)
				c.Next()
				return
			}
		}

		authorization := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Please open Linora from LINE to continue."})
			return
		}
		userID, err := verifier.VerifyIDToken(c.Request.Context(), strings.TrimSpace(authorization[len("Bearer "):]))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Your LINE session has expired. Please open Linora from LINE again."})
			return
		}
		c.Set(lineUserIDKey, userID)
		c.Next()
	}
}

func LineUserID(c *gin.Context) string {
	value, _ := c.Get(lineUserIDKey)
	userID, _ := value.(string)
	return userID
}
