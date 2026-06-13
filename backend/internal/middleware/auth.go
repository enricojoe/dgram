package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/enricojoe/dgram/backend/internal/util"
)

// contextUserID is the gin context key under which the authenticated user id is
// stored.
const contextUserID = "userID"

// RequireAuth returns middleware that validates a Bearer access token and stores
// the resolved user id in the context, aborting with 401 on any failure.
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			util.Fail(c, http.StatusUnauthorized, "missing or malformed Authorization header")
			c.Abort()
			return
		}

		token := strings.TrimSpace(header[len(prefix):])
		userID, err := util.ParseToken(token, secret, util.TokenAccess)
		if err != nil {
			util.Fail(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(contextUserID, userID)
		c.Next()
	}
}

// UserID returns the authenticated user id stored by RequireAuth, if present.
func UserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(contextUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
