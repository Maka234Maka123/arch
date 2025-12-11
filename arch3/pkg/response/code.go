package response

// 业务错误码定义
//
// 6位数错误码规范：AABBCC
//   - AA: 大类 (10-通用, 20-用户, 30-业务, 90-系统)
//   - BB: 小类 (01-09)
//   - CC: 具体错误 (01-99)
//
// 设计原则：
//   1. 前端可根据大类做统一处理（如 10xxxx 跳登录页）
//   2. 后端可根据小类做日志分级
//   3. 具体码用于精确的错误提示

const (
	CodeSuccess = 0 // 成功

	// ========== 10xxxx: 通用错误 ==========

	// 1001xx: 认证相关 - 需要重新登录
	CodeUnauthorized   = 100101 // 未认证
	CodeTokenExpired   = 100102 // Token已过期
	CodeTokenInvalid   = 100103 // Token无效
	CodeSessionExpired = 100104 // 会话已过期
	CodeLoginConflict  = 100105 // 账号在其他设备登录

	// 1002xx: 权限相关 - 无需重新登录，提示无权限
	CodeForbidden      = 100201 // 无权限
	CodeRoleRequired   = 100202 // 需要特定角色
	CodeIPBlocked      = 100203 // IP被封禁
	CodeAccountLocked  = 100204 // 账号已锁定

	// 1003xx: 请求参数 - 客户端参数问题
	CodeBadRequest   = 100301 // 请求参数错误
	CodeValidation   = 100302 // 参数校验失败
	CodeMissingParam = 100303 // 缺少必需参数
	CodeInvalidParam = 100304 // 参数格式或值无效

	// 1004xx: 资源相关 - 资源状态问题
	CodeNotFound      = 100401 // 资源不存在
	CodeAlreadyExists = 100402 // 资源已存在
	CodeConflict      = 100403 // 资源冲突

	// 1005xx: 限流相关 - 请求频率问题
	CodeTooManyRequests = 100501 // 请求过于频繁
	CodeQuotaExceeded   = 100502 // 配额已用完

	// ========== 20xxxx: 用户相关错误 ==========

	// 2001xx: 登录注册
	CodeLoginFailed     = 200101 // 登录失败
	CodePasswordError   = 200102 // 密码错误
	CodeCaptchaError    = 200103 // 验证码错误
	CodeCaptchaExpired  = 200104 // 验证码已过期
	CodeUserNotFound    = 200105 // 用户不存在
	CodeUserDisabled    = 200106 // 用户已禁用
	CodePhoneRegistered = 200107 // 手机号已注册
	CodeEmailRegistered = 200108 // 邮箱已注册

	// ========== 30xxxx: 业务相关错误 ==========

	// 3001xx: 短信相关
	CodeSMSSendFailed    = 300101 // 短信发送失败
	CodeSMSTooFrequent   = 300102 // 短信发送太频繁（1分钟限制）
	CodeSMSDailyLimit    = 300103 // 短信日发送量已达上限
	CodeSMSCodeInvalid   = 300104 // 短信验证码错误或已过期（统一错误码，避免信息泄露）
	CodeSMSVerifyTooMany = 300105 // 短信验证码验证次数过多

	// 3002xx: 文件相关
	CodeFileNotFound  = 300201 // 文件不存在
	CodeFileTooLarge  = 300202 // 文件过大
	CodeFileTypeError = 300203 // 文件类型错误
	CodeUploadFailed  = 300204 // 上传失败

	// 3003xx: 订单相关（按需启用）
	// CodeOrderNotFound = 300301
	// CodeOrderPaid     = 300302

	// 3004xx: 支付相关（按需启用）
	// CodePaymentFailed = 300401
	// CodeRefundFailed  = 300402

	// ========== 90xxxx: 系统错误 ==========

	// 9001xx: 服务器内部错误 - 需要开发排查
	CodeInternal      = 900101 // 服务器内部错误
	CodeDatabaseError = 900102 // 数据库错误
	CodeCacheError    = 900103 // 缓存错误
	CodeTimeout       = 900104 // 请求超时

	// 9002xx: 服务状态 - 可能需要运维处理
	CodeUnavailable = 900201 // 服务不可用
	CodeMaintenance = 900202 // 服务维护中
	CodeOverload    = 900203 // 服务过载

	// 9003xx: 第三方服务异常
	CodeThirdPartyError = 900301 // 第三方服务错误
	CodeOSSError        = 900302 // 存储服务异常
	CodePayChannelError = 900303 // 支付渠道异常
)

// ========== 错误码分类判断 ==========

// IsAuthError 是否认证错误（需要重新登录）
func IsAuthError(code int) bool {
	return code >= 100100 && code < 100200
}

// IsPermissionError 是否权限错误
func IsPermissionError(code int) bool {
	return code >= 100200 && code < 100300
}

// IsClientError 是否客户端错误（参数、资源、限流）
func IsClientError(code int) bool {
	return code >= 100300 && code < 100600
}

// IsUserError 是否用户业务错误
func IsUserError(code int) bool {
	return code >= 200000 && code < 300000
}

// IsBizError 是否业务错误
func IsBizError(code int) bool {
	return code >= 300000 && code < 400000
}

// IsSystemError 是否系统错误
func IsSystemError(code int) bool {
	return code >= 900000 && code < 1000000
}

// NeedsRelogin 是否需要重新登录
func NeedsRelogin(code int) bool {
	return IsAuthError(code)
}
