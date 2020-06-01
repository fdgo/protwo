package rsp

import (
	"github.com/gin-gonic/gin"
)

type ErrorSrvCall struct {
	Id     string `json:"id"`
	Code   int    `json:"code"`
	Detail string `json:"detail"`
	Status string `json:"status"`
}

type Result struct {
	Code      int32       `json:"code"`
	ClientMsg string      `json:"clientcode"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
}

func newResult() *Result {
	return &Result{
		Code:      200,
		ClientMsg: "Default ClientMsg",
		Msg:       "Default Msg",
		Data:      nil,
	}
}
func RespGin(httpCode int32, innerCode int32, clientmsg string, InnerMsg string, data interface{}, c *gin.Context) {
	resutl := newResult()
	resutl.Code = innerCode
	resutl.ClientMsg = clientmsg
	resutl.Msg = InnerMsg
	resutl.Data = data
	c.JSON(
		int(httpCode),
		resutl,
	)
}
