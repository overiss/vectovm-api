package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/overiss/vectovm-api/internal/auth"
	authservice "github.com/overiss/vectovm-api/internal/service/auth"
	"github.com/overiss/vectovm-api/internal/model"
)

const OAuthUserIDKey = "oauth_user_id"

func Auth(verifier *auth.Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := authservice.BearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{Error: "missing bearer token"})
			return
		}

		claims, err := verifier.Verify(c.Request.Context(), token)
		if err != nil {
			status := http.StatusUnauthorized
			message := "invalid token"
			if errors.Is(err, auth.ErrTokenExpired) {
				message = "token expired"
			}
			c.AbortWithStatusJSON(status, model.ErrorResponse{Error: message})
			return
		}

		oauthUserID, err := uuid.Parse(claims.Subject)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{Error: "invalid token subject"})
			return
		}

		c.Set(OAuthUserIDKey, oauthUserID)
		c.Next()
	}
}

func OAuthUserID(c *gin.Context) (uuid.UUID, bool) {
	value, ok := c.Get(OAuthUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := value.(uuid.UUID)
	return id, ok
}
