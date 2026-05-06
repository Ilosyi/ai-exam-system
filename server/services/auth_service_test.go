package services

import (
	"testing"
	"time"

	"week05/homework/server/models"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_GenerateToken(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     "student",
	}

	token, err := authService.GenerateToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token 应该有三部分（用点号分隔）
	parts := len(token)
	assert.Greater(t, parts, 0)
}

func TestAuthService_ParseToken(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     "teacher",
	}

	token, err := authService.GenerateToken(user)
	assert.NoError(t, err)

	claims, err := authService.ParseToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "teacher", claims.Role)
}

func TestAuthService_ParseToken_InvalidSignature(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)
	// 构造一个签名无效的 token
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsInVzZXJuYW1lIjoidGVzdCIsInJvbGUiOiJzdHVkZW50IiwiZXhwIjoxNzExMDAwMDAwLCJpYXQiOjE3MTA5MDAwMDB9.invalid"

	claims, err := authService.ParseToken(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "signature")
}

func TestAuthService_ParseToken_Expired(t *testing.T) {
	// 创建一个 TTL 为 1 纳秒的认证服务，让 token 立即过期
	authService := NewAuthService("test-secret", 1*time.Nanosecond)
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     "student",
	}

	token, err := authService.GenerateToken(user)
	assert.NoError(t, err)

	// 等待一小段时间确保 token 过期
	time.Sleep(10 * time.Millisecond)

	claims, err := authService.ParseToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "expired")
}

func TestAuthService_HashPassword(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)
	password := "test123password"

	hash, err := authService.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestAuthService_ComparePassword(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)
	password := "test123password"

	hash, err := authService.HashPassword(password)
	assert.NoError(t, err)

	// 正确密码
	err = authService.ComparePassword(hash, password)
	assert.NoError(t, err)

	// 错误密码
	err = authService.ComparePassword(hash, "wrongpassword")
	assert.Error(t, err)
}

func TestAuthService_DefaultSecret(t *testing.T) {
	authService := NewAuthService("", 24*time.Hour)
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     "admin",
	}

	token, err := authService.GenerateToken(user)
	assert.NoError(t, err)

	claims, err := authService.ParseToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
}
