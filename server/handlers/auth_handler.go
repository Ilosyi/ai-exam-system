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

type AuthHandler struct {
	userRepo    *repositories.UserRepository
	authService *services.AuthService
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Role     string `json:"role" binding:"omitempty,oneof=teacher student"`
	ClassID  *uint  `json:"classId"`
}

type refreshRequest struct {
	Token string `json:"token" binding:"required"`
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,min=6,max=64"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=64"`
}

type updateUserRequest struct {
	Role    string `json:"role" binding:"omitempty,oneof=admin teacher student"`
	ClassID *uint  `json:"classId"`
	Status  string `json:"status" binding:"omitempty,oneof=active disabled"`
}

type authUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	ClassID  *uint  `json:"classId"`
	Status   string `json:"status"`
}

type authPayload struct {
	Token string           `json:"token"`
	User  authUserResponse `json:"user"`
}

func NewAuthHandler(userRepo *repositories.UserRepository, authService *services.AuthService) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

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
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"message": "账号已停用"})
		return
	}
	if err := h.authService.ComparePassword(user.PasswordHash, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
		return
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "生成 token 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": authPayload{
			Token: token,
			User:  toAuthUserResponse(*user),
		},
	})
}

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

	if _, err := h.userRepo.FindByUsername(context.Background(), username); err == nil {
		c.JSON(http.StatusConflict, gin.H{"message": "用户名已存在"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询用户失败"})
		return
	}

	role := req.Role
	if role == "" {
		role = "student"
	}

	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "密码加密失败"})
		return
	}

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

func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未登录"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toAuthUserResponse(*user)})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	claims, err := h.authService.ParseToken(strings.TrimSpace(req.Token))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "token 无效或已过期"})
		return
	}

	user, err := h.userRepo.FindByID(context.Background(), claims.UserID)
	if err != nil || user.Status != "active" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户不存在或已停用"})
		return
	}

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

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "退出成功"})
}

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

	fullUser, err := h.userRepo.FindByID(context.Background(), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "用户不存在"})
		return
	}
	if err := h.authService.ComparePassword(fullUser.PasswordHash, req.OldPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "旧密码错误"})
		return
	}

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

func (h *AuthHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的用户ID"})
		return
	}

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

func toAuthUserResponse(user models.User) authUserResponse {
	return authUserResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		ClassID:  user.ClassID,
		Status:   user.Status,
	}
}
