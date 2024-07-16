package biz

const (
	CacheNSForUser = "user"
	CacheNSForRole = "role"
)

const (
	CacheKeyForSyncToCasbin = "sync:casbin"
)

const (
	ErrInvalidTokenID            = "com.invalid.token"
	ErrInvalidCaptchaID          = "com.invalid.captcha"
	ErrInvalidUsernameOrPassword = "com.invalid.username-or-password"
)

const (
	disableCompression = "disableCompression"
	disableKeepAlive   = "disableKeepAlive"
	disableRedirects   = "disableRedirect"
	h2                 = "enableHttp2"
)

const (
	maxTimeoutSecond = 5
	maxTaskCount     = 50
)

const (
	maxDynamicParamLength int = 1e5
)

const (
	PopWaitSecond = 10
)
