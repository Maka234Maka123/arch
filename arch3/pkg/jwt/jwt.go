// Package jwt 提供 JWT 认证相关的功能
// 包括 token 生成、解析、验证和 cookie 管理
package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// Cookie 键名
	AccessTokenCookieKey  = "ap-access-token"
	RefreshTokenCookieKey = "ap-refresh-token"

	// Token 类型
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	// Token 过期时间默认值
	DefaultAccessTokenDuration  = 15 * time.Minute   // 默认 Access Token: 15分钟
	DefaultRefreshTokenDuration = 7 * 24 * time.Hour // 默认 Refresh Token: 7天
)

// Config JWT 配置
type Config struct {
	Secret        string        // JWT 签名密钥
	AccessExpire  time.Duration // Access Token 过期时间
	RefreshExpire time.Duration // Refresh Token 过期时间
	CookieSecure  bool          // Cookie 是否仅通过 HTTPS 传输
}

// Claims JWT claims 结构
type Claims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"` // access 或 refresh
	jwt.RegisteredClaims
}

// TokenPair 访问令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Manager JWT 管理器
type Manager struct {
	secret        []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
	cookieSecure  bool
	rdb           *redis.Client
}

// NewManager 创建 JWT 管理器
func NewManager(cfg *Config, rdb *redis.Client) *Manager {
	accessExpire := DefaultAccessTokenDuration
	refreshExpire := DefaultRefreshTokenDuration

	if cfg.AccessExpire > 0 {
		accessExpire = cfg.AccessExpire
	}
	if cfg.RefreshExpire > 0 {
		refreshExpire = cfg.RefreshExpire
	}

	return &Manager{
		secret:        []byte(cfg.Secret),
		accessExpire:  accessExpire,
		refreshExpire: refreshExpire,
		cookieSecure:  cfg.CookieSecure,
		rdb:           rdb,
	}
}

// GenerateTokenPair 生成访问令牌对（短 token + 长 token）
func (m *Manager) GenerateTokenPair(userID string) (*TokenPair, error) {
	now := time.Now()

	// 生成 Access Token (短 token)
	accessJTI := uuid.New().String()
	accessClaims := &Claims{
		UserID:    userID,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessJTI,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessExpire)),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// 生成 Refresh Token (长 token)
	refreshJTI := uuid.New().String()
	refreshClaims := &Claims{
		UserID:    userID,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshExpire)),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

// ParseToken 解析并验证 token
func (m *Manager) ParseToken(tokenString string, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// 验证 token 类型
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("expected %s token, got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}

// SetTokensInCookie 将 token 对设置到 cookie 中
func (m *Manager) SetTokensInCookie(c *app.RequestContext, tokenPair *TokenPair) {
	domain := getSecondaryDomain(string(c.Host()))

	// 设置 Access Token
	c.SetCookie(
		AccessTokenCookieKey,
		tokenPair.AccessToken,
		int(m.accessExpire.Seconds()),
		"/",
		domain,
		protocol.CookieSameSiteStrictMode,
		m.cookieSecure, // secure - 由配置控制
		true,           // httpOnly
	)

	// 设置 Refresh Token
	c.SetCookie(
		RefreshTokenCookieKey,
		tokenPair.RefreshToken,
		int(m.refreshExpire.Seconds()),
		"/",
		domain,
		protocol.CookieSameSiteStrictMode,
		m.cookieSecure, // secure - 由配置控制
		true,           // httpOnly
	)
}

// ClearTokensFromCookie 清除 cookie 中的 token
func (m *Manager) ClearTokensFromCookie(c *app.RequestContext) {
	domain := getSecondaryDomain(string(c.Host()))

	// 清除 Access Token
	c.SetCookie(
		AccessTokenCookieKey,
		"",
		-1,
		"/",
		domain,
		protocol.CookieSameSiteStrictMode,
		m.cookieSecure,
		true,
	)

	// 清除 Refresh Token
	c.SetCookie(
		RefreshTokenCookieKey,
		"",
		-1,
		"/",
		domain,
		protocol.CookieSameSiteStrictMode,
		m.cookieSecure,
		true,
	)
}

// IsTokenBlacklisted 检查 token 是否在黑名单中
func (m *Manager) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("token:blacklist:%s", jti)
	val, err := m.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return val == "1", nil
}

// AddTokenToBlacklist 将 token 加入黑名单
func (m *Manager) AddTokenToBlacklist(ctx context.Context, jti string, expiration time.Duration) error {
	key := fmt.Sprintf("token:blacklist:%s", jti)
	return m.rdb.Set(ctx, key, "1", expiration).Err()
}

// RevokeRefreshToken 撤销 refresh token（用于 token 轮转）
func (m *Manager) RevokeRefreshToken(ctx context.Context, jti string) error {
	return m.AddTokenToBlacklist(ctx, jti, m.refreshExpire)
}

// GetAccessExpire 获取 access token 过期时间
func (m *Manager) GetAccessExpire() time.Duration {
	return m.accessExpire
}

// GetRefreshExpire 获取 refresh token 过期时间
func (m *Manager) GetRefreshExpire() time.Duration {
	return m.refreshExpire
}

// 多级顶级域名后缀
var multiLevelTLDs = map[string]bool{
	"com.cn": true,
}

// getSecondaryDomain 获取二级域名用于 cookie domain
// 例如: api.example.com -> .example.com
// 例如: api.example.co.uk -> .example.co.uk
func getSecondaryDomain(host string) string {
	// 移除端口
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// 本地开发环境
	if host == "localhost" || host == "127.0.0.1" {
		return ""
	}

	// 分割域名
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return host
	}

	// 检查是否是多级顶级域名
	if len(parts) >= 3 {
		possibleTLD := parts[len(parts)-2] + "." + parts[len(parts)-1]
		if multiLevelTLDs[possibleTLD] {
			// 需要取最后三段: example.co.uk
			if len(parts) >= 3 {
				return "." + strings.Join(parts[len(parts)-3:], ".")
			}
		}
	}

	// 普通域名取最后两段: example.com
	return "." + strings.Join(parts[len(parts)-2:], ".")
}
