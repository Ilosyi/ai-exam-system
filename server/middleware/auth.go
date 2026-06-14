// ============================================================================
// middleware/auth.go - 认证与鉴权中间件
// ============================================================================
//
// 本文件实现了两个 Gin 中间件：
// 1. RequireAuth - 认证中间件：验证 JWT 令牌，确保用户已登录
// 2. RequireRoles - 鉴权中间件：检查用户角色，确保有权限访问
//
// 中间件是什么？
// 中间件是位于"请求到达路由处理器之前"的一层拦截器。
// 它可以：
// - 检查请求是否合法（如是否携带有效的 JWT）
// - 修改请求上下文（如将用户信息写入 ctx）
// - 终止请求（如返回 401 未授权）
//
// 请求处理流程：
//   客户端请求 → CORS 中间件 → Auth 中间件 → Role 中间件 → Handler
//
// 学习要点：
// - Gin 中间件的工作原理（c.Next()、c.AbortWithStatusJSON()）
// - 如何在中间件之间传递数据（c.Set()、c.Get()）
// - 认证（你是谁）vs 鉴权（你能做什么）
// ============================================================================

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"week05/homework/server/models"
	"week05/homework/server/repositories"
	"week05/homework/server/services"
)

// CurrentUserKey 是 Gin Context 中存储当前用户的键名。
// 使用常量避免拼写错误，也便于在多个地方引用。
const CurrentUserKey = "currentUser"

// RequireAuth 创建一个认证中间件。
//
// 这个中间件会：
// 1. 从请求头中提取 Authorization 字段
// 2. 验证 Bearer token 格式
// 3. 解析 JWT 令牌，提取用户信息
// 4. 查询数据库确认用户存在且状态为 active
// 5. 将用户信息写入 Gin Context（后续的 Handler 可以读取）
//
// 如果任何一步失败，返回 401 Unauthorized 并终止请求。
//
// 参数：
//   - authService: 认证服务，用于解析 JWT
//   - userRepo: 用户数据访问层，用于查询用户信息
//
// 返回值：
//   - gin.HandlerFunc: Gin 中间件函数
//
// 使用示例：
//
//	protected := api.Group("")
//	protected.Use(middleware.RequireAuth(authService, userRepo))
//	protected.GET("/me", handler.Me)
func RequireAuth(authService *services.AuthService, userRepo *repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 第一步：从请求头提取 Authorization 字段
		// 格式应为："Bearer <token>"
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			// 如果没有 Authorization 头或格式不对，返回 401
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "未登录或 token 缺失"})
			return
		}

		// 第二步：提取 token 部分（去掉 "Bearer " 前缀）
		token := strings.TrimSpace(authHeader[len("Bearer "):])

		// 第三步：解析 JWT 令牌
		// ParseToken 会验证签名、检查过期时间，并返回用户信息（Claims）
		claims, err := authService.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "token 无效或已过期"})
			return
		}

		// 第四步：查询数据库确认用户存在
		// 即使 JWT 解析成功，也需要确认用户没有被删除或停用
		user, err := userRepo.FindByID(c.Request.Context(), claims.UserID)
		if err != nil || user.Status != "active" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "用户不存在或已停用"})
			return
		}

		// 第五步：将用户信息写入 Gin Context
		// 后续的 Handler 可以通过 GetCurrentUser(c) 读取
		c.Set(CurrentUserKey, *user)

		// 继续处理后续的中间件和路由处理器
		c.Next()
	}
}

// RequireRoles 创建一个鉴权中间件。
//
// 这个中间件会检查当前用户的角色是否在允许的角色列表中。
// 它必须在 RequireAuth 之后使用（因为需要先有用户信息）。
//
// 参数：
//   - roles: 允许访问的角色列表，如 "admin", "teacher", "student"
//
// 返回值：
//   - gin.HandlerFunc: Gin 中间件函数
//
// 使用示例：
//
//	adminRoutes.Use(middleware.RequireRoles("admin"))           // 仅管理员
//	teacherRoutes.Use(middleware.RequireRoles("admin", "teacher")) // 管理员和教师
func RequireRoles(roles ...string) gin.HandlerFunc {
	// 将角色列表转为 map，方便快速查找
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		// 从 Context 获取当前用户（由 RequireAuth 中间件写入）
		user, ok := GetCurrentUser(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
			return
		}

		// 检查用户角色是否在允许列表中
		if _, exists := allowed[user.Role]; !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "无权限访问当前资源"})
			return
		}

		// 角色验证通过，继续处理
		c.Next()
	}
}

// ---- 辅助函数：从 Gin Context 中提取当前用户信息 ----

// GetCurrentUser 从 Gin Context 中获取当前登录用户。
//
// 这个函数通常在 Handler 中调用，用于获取由 RequireAuth 中间件写入的用户信息。
//
// 返回值：
//   - *models.User: 用户指针，如果未登录则为 nil
//   - bool: 是否成功获取到用户
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	// c.Get 从 Context 中读取值
	raw, ok := c.Get(CurrentUserKey)
	if !ok {
		return nil, false
	}
	// 类型断言：将 interface{} 转换为 models.User
	user, ok := raw.(models.User)
	if !ok {
		return nil, false
	}
	return &user, true
}

// GetCurrentUserID 从 Gin Context 中获取当前用户的 ID。
//
// 这是一个便捷函数，内部调用 GetCurrentUser。
// 如果未登录，返回 0（Go 的 uint 零值）。
func GetCurrentUserID(c *gin.Context) uint {
	user, ok := GetCurrentUser(c)
	if !ok {
		return 0
	}
	return user.ID
}

// GetCurrentUserRole 从 Gin Context 中获取当前用户的角色。
//
// 这是一个便捷函数，内部调用 GetCurrentUser。
// 如果未登录，返回空字符串。
func GetCurrentUserRole(c *gin.Context) string {
	user, ok := GetCurrentUser(c)
	if !ok {
		return ""
	}
	return user.Role
}
