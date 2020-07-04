package api

const (
	successCode = 0
	failedCode  = 1
)

const (
	OK            = "ok"
	ErrParameters = "参数错误"
)

type resultInfo struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
