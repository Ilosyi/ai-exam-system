// ============================================================================
// handlers/auth_handler.go - 认证接口处理器
// ============================================================================
//
// 本文件实现了用户认证相关的 HTTP 接口，包括：
// - Login:          用户登录（验证用户名密码，返回 JWT）
// - Register:       用户注册（创建新用户）
// - Me:             获取当前登录用户信息
// - Refresh:        刷新 JWT 令牌
// - Logout:         退出登录
// - ChangePassword: 修改密码
// - ListUsers:      用户列表（管理员）
// - UpdateUser:     更新用户信息（管理员）
// - DeleteUser:     删除用户（管理员）
//
// Handler 的职责：
// 1. 解析请求参数（路径参数、查询参数、请求体）
// 2. 校验参数合法性
// 3. 调用 Repository/Service 完成业务逻辑
// 4. 组织响应并返回给客户端
//
// 学习要点：
// - Gin 的请求参数绑定（ShouldBindJSON、Query、Param）
// - HTTP 状态码的正确使用
// - 错误处理的最佳实践
// ============================================================================

package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"week05/homework/server/middleware"
	"week05/homework/server/models"
	"week05/homework/server/repositories"
	"week05/homework/server/services"
)

// AuthHandler 处理认证相关的 HTTP 请求。
//
// 依赖：
// - userRepo: 用户数据访问层（查询、创建、更新用户）
// - authService: 认证服务（JWT、密码哈希）
type AuthHandler struct {
	userRepo    *repositories.UserRepository
	authService *services.AuthService
}

// ---- 请求体结构体 ----
// 使用 Gin 的 binding 标签进行参数校验

// loginRequest 登录请求体
type loginRequest struct {
	Username string `json:"username" binding:"required"`             // 用户名（必填）
	Password string `json:"password" binding:"required,min=6"`      // 密码（必填，最少 6 位）
}

// registerRequest 注册请求体
type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"` // 用户名（3-32 字符）
	Password string `json:"password" binding:"required,min=6,max=64"` // 密码（6-64 字符）
	Role     string `json:"role" binding:"omitempty,oneof=teacher student"` // 角色（可选，默认 student）
	ClassID  *uint  `json:"classId"`                                    // 班级 ID（可选）
}

// refreshRequest 刷新令牌请求体
type refreshRequest struct {
	Token string `json:"token" binding:"required"` // 旧的 JWT 令牌
}

// changePasswordRequest 修改密码请求体
type changePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,min=6,max=64"` // 旧密码
	NewPassword string `json:"newPassword" binding:"required,min=6,max=64"` // 新密码
}

// updateUserRequest 更新用户请求体
type updateUserRequest struct {
	Role    string `json:"role" binding:"omitempty,oneof=admin teacher student"` // 角色
	ClassID *uint  `json:"classId"`                                              // 班级 ID
	Status  string `json:"status" binding:"omitempty,oneof=active disabled"`    // 状态
}

// ---- 响应体结构体 ----

// authUserResponse 返回给前端的用户信息（不包含密码）
type authUserResponse struct {
	ID       uint   `json:"id"`       // 用户 ID
	Username string `json:"username"` // 用户名
	Role     string `json:"role"`     // 角色
	ClassID  *uint  `json:"classId"`  // 班级 ID
	Status   string `json:"status"`   // 状态
}

// authPayload 认证成功的响应体（包含 token 和用户信息）
type authPayload struct {
	Token string           `json:"token"` // JWT 令牌
	User  authUserResponse `json:"user"`  // 用户信息
}

// NewAuthHandler 创建一个新的 AuthHandler 实例。
func NewAuthHandler(userRepo *repositories.UserRepository, authService *services.AuthService) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, authService: authService}
}

// Login 处理用户登录请求。
//
// 流程：
// 1. 解析并校验请求参数
// 2. 根据用户名查询用户
// 3. 检查用户状态
// 4. 验证密码
// 5. 生成 JWT 令牌
// 6. 返回 token 和用户信息
//
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	// 查询用户
	user, err := h.userRepo.FindByUsername(context.Background(), strings.TrimSpace(req.Username))
	if err != nil {
		status := http.StatusInternalServerError
		message := "登录失败"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusUnauthorized
			message = "用户名或密码错误"
		}
		c.JSON(status, gin.H{"message": message})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"message": "账号已停用"})
		return
	}

	// 验证密码
	if err := h.authService.ComparePassword(user.PasswordHash, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
		return
	}

	// 生成 JWT 令牌
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "生成 token 失败"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"data": authPayload{
			Token: token,
			User:  toAuthUserResponse(*user),
		},
	})
}

// Register 处理用户注册请求。
//
// 流程：
// 1. 解析并校验请求参数
// 2. 检查用户名是否已存在
// 3. 哈希密码
// 4. 创建用户
// 5. 生成 JWT 令牌
// 6. 返回 token 和用户信息
//
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "用户名不能为空"})
		return
	}

	// 检查用户名是否已存在
	if _, err := h.userRepo.FindByUsername(context.Background(), username); err == nil {
		c.JSON(http.StatusConflict, gin.H{"message": "用户名已存在"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询用户失败"})
		return
	}

	// 设置默认角色
	role := req.Role
	if role == "" {
		role = "student"
	}

	// 哈希密码
	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "密码加密失败"})
		return
	}

	// 创建用户
	user := models.User{
		Username:     username,
		Role:         role,
		ClassID:      req.ClassID,
		PasswordHash: passwordHash,
		Status:       "active",
	}
	if err := h.userRepo.Create(context.Background(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "注册失败", "error": err.Error()})
		return
	}

	// 生成 JWT 令牌
	token, err := h.authService.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "生成 token 失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": authPayload{
			Token: token,
			User:  toAuthUserResponse(user),
		},
	})
}

// Me 获取当前登录用户信息。
//
// GET /api/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toAuthUserResponse(*user)})
}

// Refresh 刷新 JWT 令牌。
//
// 流程：
// 1. 解析旧 token
// 2. 查询用户确认存在且活跃
// 3. 生成新 token
//
// POST /api/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	// 解析旧 token
	claims, err := h.authService.ParseToken(strings.TrimSpace(req.Token))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "token 无效或已过期"})
		return
	}

	// 查询用户
	user, err := h.userRepo.FindByID(context.Background(), claims.UserID)
	if err != nil || user.Status != "active" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户不存在或已停用"})
		return
	}

	// 生成新 token
	newToken, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "刷新 token 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": authPayload{
			Token: newToken,
			User:  toAuthUserResponse(*user),
		},
	})
}

// Logout 退出登录。
//
// 注意：由于 JWT 是无状态的，服务端不需要做任何清理。
// 前端只需要清除本地存储的 token 即可。
//
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "退出成功"})
}

// ChangePassword 修改密码。
//
// 流程：
// 1. 验证旧密码
// 2. 哈希新密码
// 3. 更新数据库
//
// POST /api/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}
	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"message": "新旧密码不能一致"})
		return
	}

	// 查询完整用户信息（包含密码哈希）
	fullUser, err := h.userRepo.FindByID(context.Background(), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "用户不存在"})
		return
	}

	// 验证旧密码
	if err := h.authService.ComparePassword(fullUser.PasswordHash, req.OldPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "旧密码错误"})
		return
	}

	// 哈希新密码并更新
	hash, err := h.authService.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "密码加密失败"})
		return
	}

	if err := h.userRepo.Update(context.Background(), user.ID, map[string]interface{}{"password_hash": hash}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "修改密码失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// ListUsers 获取用户列表（管理员）。
//
// GET /api/users
func (h *AuthHandler) ListUsers(c *gin.Context) {
	page := parseIntWithDefault(c.Query("page"), 1)
	pageSize := parseIntWithDefault(c.Query("pageSize"), 10)
	filters := repositories.UserFilters{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Role:     strings.TrimSpace(c.Query("role")),
		Status:   strings.TrimSpace(c.Query("status")),
		Page:     page,
		PageSize: pageSize,
	}
	if classVal := strings.TrimSpace(c.Query("classId")); classVal != "" {
		id, err := strconv.Atoi(classVal)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "classId 参数无效"})
			return
		}
		uid := uint(id)
		filters.ClassID = &uid
	}

	items, total, err := h.userRepo.List(context.Background(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "error": err.Error()})
		return
	}

	resp := make([]authUserResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toAuthUserResponse(item))
	}

	c.JSON(http.StatusOK, gin.H{"data": resp, "total": total})
}

// UpdateUser 更新用户信息（管理员）。
//
// PUT /api/users/:id
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的用户ID"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.ClassID != nil {
		updates["class_id"] = *req.ClassID
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无更新内容"})
		return
	}

	if err := h.userRepo.Update(context.Background(), uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新失败", "error": err.Error()})
		return
	}

	updated, err := h.userRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": toAuthUserResponse(*updated)})
}

// DeleteUser 删除用户（管理员）。
//
// DELETE /api/users/:id
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的用户ID"})
		return
	}

	// 防止删除自己
	currentID := middleware.GetCurrentUserID(c)
	if uint(id) == currentID {
		c.JSON(http.StatusBadRequest, gin.H{"message": "不能删除当前登录账号"})
		return
	}

	if err := h.userRepo.Delete(context.Background(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除失败", "error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// toAuthUserResponse 将 User 模型转换为响应结构体（去除敏感信息）。
func toAuthUserResponse(user models.User) authUserResponse {
	return authUserResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		ClassID:  user.ClassID,
		Status:   user.Status,
	}
}
