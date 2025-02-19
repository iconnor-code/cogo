package cerr

const (
	ERR_CODE_CLIENT   = 4000 // 客户端导致的错误
	ERR_CODE_INVALID_PARAM = 4001 // 客户端参数错误

	ERR_CODE_INTERNAL = 5000 // 服务端导致的错误
	ERR_CODE_INTERNAL_PANIC = 5500 // 服务端panic
	ERR_CODE_INTERNAL_ERROR = 5501 // 服务端未知错误

	ERR_CODE_EXTRA    = 6000 // 调用外部系统导致的错误

)

var (
	ErrInvalidParams    = NewCustomError(ERR_CODE_INVALID_PARAM, "invalid params", nil)

	ErrInternalPanic = NewCustomError(ERR_CODE_INTERNAL_PANIC, "internal server panic", nil)
	ErrInternalError = NewCustomError(ERR_CODE_INTERNAL_ERROR, "internal server error", nil)
)
