// ============================================================================
// services/auth_service.go - 认证服务
// ============================================================================
//
// 本文件实现了用户认证相关的业务逻辑，包括：
// 1. JWT（JSON Web Token）的生成和解析
// 2. 密码的哈希和验证
//
// JWT 是什么？
// JWT 是一种用于身份验证的令牌标准，由三部分组成：
// - Header（头部）：声明算法和令牌类型
// - Payload（载荷）：存储用户信息（如 ID、角色、过期时间）
// - Signature（签名）：用于验证令牌是否被篡改
//
// JWT 的优势：
// - 无状态：服务器不需要存储会话信息
// - 可扩展：可以在不同服务之间传递
// - 自包含：用户信息直接存储在令牌中
//
// 本项目没有使用第三方 JWT 库，而是用标准库手写实现，
// 这样可以更好地理解 JWT 的工作原理。
//
// 学习要点：
// - JWT 的结构和工作原理
// - HMAC-SHA256 签名算法
// - bcrypt 密码哈希
// - Base64 编码
// ============================================================================

package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt" // bcrypt 密码哈希库

	"week05/homework/server/models"
)

// AuthClaims 是 JWT 的载荷（Payload）部分。
//
// 它包含了用户的身份信息和令牌的元数据：
// - UserID:    用户 ID
// - Username:  用户名
// - Role:      用户角色
// - ExpiresAt: 过期时间（Unix 时间戳）
// - IssuedAt:  签发时间（Unix 时间戳）
//
// JSON 标签使用了缩写形式（如 "uid" 而不是 "userId"），
// 这是为了减小 JWT 令牌的大小。
type AuthClaims struct {
	UserID    uint   `json:"uid"`   // 用户 ID
	Username  string `json:"username"` // 用户名
	Role      string `json:"role"`    // 角色
	ExpiresAt int64  `json:"exp"`     // 过期时间（Unix 时间戳）
	IssuedAt  int64  `json:"iat"`     // 签发时间（Unix 时间戳）
}

// AuthService 提供认证相关的业务逻辑。
//
// 它持有一个密钥（secret）和令牌有效期（ttl），
// 用于 JWT 的签名和验证。
type AuthService struct {
	secret []byte       // JWT 签名密钥
	ttl    time.Duration // JWT 有效期
}

// NewAuthService 创建一个新的 AuthService 实例。
//
// 参数：
//   - secret: JWT 签名密钥（如果为空，使用默认密钥）
//   - ttl:    JWT 有效期
//
// 安全提示：生产环境应该使用强随机密钥，并通过环境变量注入。
func NewAuthService(secret string, ttl time.Duration) *AuthService {
	// 如果没有提供密钥，使用默认密钥（仅用于开发环境）
	if strings.TrimSpace(secret) == "" {
		secret = "week05-homework-dev-secret"
	}
	// 如果没有提供有效期，默认 24 小时
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &AuthService{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// GenerateToken 为用户生成 JWT 令牌。
//
// JWT 的结构：header.payload.signature
//
// 生成流程：
// 1. 创建 Header（声明算法和类型）
// 2. 创建 Payload（存储用户信息和过期时间）
// 3. 将 Header 和 Payload 分别 Base64 编码
// 4. 使用 HMAC-SHA256 对 "header.payload" 签名
// 5. 将签名 Base64 编码
// 6. 拼接为最终的 JWT 字符串
//
// 参数：
//   - user: 用户对象
//
// 返回值：
//   - string: JWT 令牌字符串
//   - error: 生成失败时返回错误
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	// 第一步：创建 Header
	headerJSON, err := json.Marshal(map[string]string{
		"alg": "HS256",  // 签名算法：HMAC-SHA256
		"typ": "JWT",    // 令牌类型：JWT
	})
	if err != nil {
		return "", err
	}

	// 第二步：创建 Payload
	now := time.Now()
	claims := AuthClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		IssuedAt:  now.Unix(),                // 签发时间
		ExpiresAt: now.Add(s.ttl).Unix(),     // 过期时间
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	// 第三步：Base64 编码 Header 和 Payload
	// 使用 RawURLEncoding（URL 安全的 Base64，不包含 + / = 字符）
	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// 第四步：签名
	// 签名内容 = header + "." + payload
	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, s.secret)
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return "", err
	}
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	// 第五步：拼接最终的 JWT
	return signingInput + "." + signature, nil
}

// ParseToken 解析并验证 JWT 令牌。
//
// 验证流程：
// 1. 拆分 JWT 为三部分（header.payload.signature）
// 2. 使用密钥重新计算签名，与传入的签名对比
// 3. 解析 Payload 中的 Claims
// 4. 检查令牌是否过期
//
// 参数：
//   - token: JWT 令牌字符串
//
// 返回值：
//   - *AuthClaims: 解析出的用户信息
//   - error: 验证失败时返回错误（格式错误、签名无效、已过期等）
func (s *AuthService) ParseToken(token string) (*AuthClaims, error) {
	// 第一步：拆分 JWT
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("token format invalid")
	}

	// 第二步：验证签名
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, s.secret)
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return nil, err
	}
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature failed: %w", err)
	}
	// hmac.Equal 是常量时间比较，防止时序攻击
	if !hmac.Equal(actualSig, expectedSig) {
		return nil, errors.New("token signature invalid")
	}

	// 第三步：解析 Payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload failed: %w", err)
	}

	var claims AuthClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("parse claims failed: %w", err)
	}

	// 第四步：验证 Claims 的基本有效性
	if claims.UserID == 0 || claims.ExpiresAt == 0 {
		return nil, errors.New("token claims invalid")
	}

	// 第五步：检查是否过期
	if time.Now().Unix() >= claims.ExpiresAt {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

// HashPassword 使用 bcrypt 算法对密码进行哈希。
//
// bcrypt 是一种专门用于密码哈希的算法，具有以下特点：
// - 单向性：无法从哈希值反推出原始密码
// - 慢速性：故意设计得很慢，增加暴力破解的难度
// - 加盐：自动添加随机盐值，防止彩虹表攻击
//
// 参数：
//   - password: 明文密码
//
// 返回值：
//   - string: 哈希后的密码
//   - error: 哈希失败时返回错误
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ComparePassword 验证明文密码是否与哈希匹配。
//
// 参数：
//   - hashedPassword: 哈希后的密码（存储在数据库中）
//   - plainPassword:  用户输入的明文密码
//
// 返回值：
//   - error: 如果密码不匹配，返回错误；如果匹配，返回 nil
func (s *AuthService) ComparePassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
