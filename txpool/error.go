package txpool

import "errors"

var (
	errtx         = errors.New("tx is error")
	errtomuch     = errors.New("recv tx to much,so refused")
	errtxoutrange = errors.New("txpoll tx out of range,so refused")
)
