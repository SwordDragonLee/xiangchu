package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiangchu/server/internal/pkg/errs"
	"github.com/xiangchu/server/internal/pkg/jwtx"
	"github.com/xiangchu/server/internal/pkg/response"
)

const (
	CtxUID  = "uid"
	CtxRole = "role"
)

// roleAllows: required role -> allowed token lines
var roleAllows = map[string]map[string]bool{
	"user":     {"user": true, "chef": true, "supplier": true},
	"chef":     {"chef": true},
	"supplier": {"supplier": true},
	"admin":    {"admin": true},
}

// Auth validates the JWT and checks the required role.
func Auth(maker *jwtx.Maker, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		token := strings.TrimPrefix(raw, "Bearer ")
		if token == "" || raw == token {
			response.Fail(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		claims, err := maker.Parse(token)
		if err != nil {
			response.Fail(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		allowed := false
		for _, r := range roles {
			if roleAllows[r][claims.Role] {
				allowed = true
				break
			}
		}
		if !allowed {
			response.Fail(c, errs.ErrForbidden)
			c.Abort()
			return
		}
		c.Set(CtxUID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// UID reads the logged-in user id injected by Auth.UID helper
func UID(c *gin.Context) int64 {
	if v, ok := c.Get(CtxUID); ok {
		if id, ok2 := v.(int64); ok2 {
			return id
		}
	}
	return 0
}

// Role reads the token role injected by Auth.
func Role(c *gin.Context) string {
	return c.GetString(CtxRole)
}
