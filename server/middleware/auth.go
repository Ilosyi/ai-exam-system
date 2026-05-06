package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"week05/homework/server/models"
	"week05/homework/server/repositories"
	"week05/homework/server/services"
)

const CurrentUserKey = "currentUser"

func RequireAuth(authService *services.AuthService, userRepo *repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "未登录或 token 缺失"})
			return
		}

		token := strings.TrimSpace(authHeader[len("Bearer "):])
		claims, err := authService.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "token 无效或已过期"})
			return
		}

		user, err := userRepo.FindByID(c.Request.Context(), claims.UserID)
		if err != nil || user.Status != "active" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "用户不存在或已停用"})
			return
		}

		c.Set(CurrentUserKey, *user)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		user, ok := GetCurrentUser(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
			return
		}
		if _, exists := allowed[user.Role]; !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "无权限访问当前资源"})
			return
		}
		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	raw, ok := c.Get(CurrentUserKey)
	if !ok {
		return nil, false
	}
	user, ok := raw.(models.User)
	if !ok {
		return nil, false
	}
	return &user, true
}

func GetCurrentUserID(c *gin.Context) uint {
	user, ok := GetCurrentUser(c)
	if !ok {
		return 0
	}
	return user.ID
}

func GetCurrentUserRole(c *gin.Context) string {
	user, ok := GetCurrentUser(c)
	if !ok {
		return ""
	}
	return user.Role
}
