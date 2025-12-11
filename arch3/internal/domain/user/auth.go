package user

import "arch3/pkg/jwt"

// LoginResult 登录结果领域模型
type LoginResult struct {
	User      *User
	TokenPair *jwt.TokenPair
	IsNew     bool // true: 新注册, false: 已有用户登录
}
