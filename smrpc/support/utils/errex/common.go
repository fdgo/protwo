package errex

import (
	"newrpc/support/utils/timex"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"encoding/json"
)

var (
	NORMAL_BASE                  = 5000
	NORMAL_OK                    = NORMAL_BASE + 1
	NORMAL_FAIL                  = NORMAL_BASE + 2
	NORMAL_BAD_REQUEST           = NORMAL_BASE + 3
	NORMAL_NORMALUNAUTHORIZED    = NORMAL_BASE + 4
	NORMAL_FORBIDDEN             = NORMAL_BASE + 5
	NORMAL_NO_PERMISSION         = NORMAL_BASE + 6
	NORMAL_INTERNAL_SERVER_ERROR = NORMAL_BASE + 7
	NORMAL_INVALID_REQUEST_TYPE  = NORMAL_BASE + 8
	NORMAL_SYSTEM_ERROR          = NORMAL_BASE + 9
	NORMAL_INVALID_PARAMETER     = NORMAL_BASE + 10
	NORMAL_NOT_FOUND             = NORMAL_BASE + 11
	NORMAL_ACCESS_TOO_OFTEN      = NORMAL_BASE + 12
)

var MsgFlags = map[int]string{
	NORMAL_OK:                    "Success",
	NORMAL_FAIL:                  "fail",
	NORMAL_BAD_REQUEST:           "Invalid request",
	NORMAL_NORMALUNAUTHORIZED:    "unauthorized",
	NORMAL_FORBIDDEN:             "forbidden",
	NORMAL_NO_PERMISSION:         "no permission",
	NORMAL_INTERNAL_SERVER_ERROR: "server error",
	NORMAL_INVALID_REQUEST_TYPE:  "Illegal request",
	NORMAL_SYSTEM_ERROR:          "System error",
	NORMAL_INVALID_PARAMETER:     "Invalid parameter",
	NORMAL_NOT_FOUND:             "not found",
	NORMAL_ACCESS_TOO_OFTEN:      "Access too often",
//----------------------------------------------------//
	USER_REGIST_EXISTS: "accout exists",
	USER_REGIST_ACCLEN: "accout lenth error",
	//----------------------------------------------------//
	PRODUCT_ADD_PRICELEN: "product price error",
	PRODUCT_SELL_EMPTY:   "product is empty",
}

func RespMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[NORMAL_FAIL]
}

type errorInfo struct {
	Time     string `json:"time"`
	Alarm    string `json:"alarm"`
	Message  string `json:"message"`
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Funcname string `json:"funcname"`
}
// 发邮件
func Email (text string)  {
	alarm("EMAIL", text,2)
}
// 发短信
func Sms (text string)  {
	alarm("SMS", text, 2)
}
// 发微信
func WeChat (text string)  {
	alarm("WX", text, 2)
}
func alarm(level string, str string, skip int) {
	// 当前时间
	currentTime := timex.GetCurrentTime()

	// 定义 文件名、行号、方法名
	fileName, line, functionName := "?", 0 , "?"

	pc, fileName, line, ok := runtime.Caller(skip)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
		functionName = filepath.Ext(functionName)
		functionName = strings.TrimPrefix(functionName, ".")
	}
	var msg = errorInfo {
		Time     : currentTime,
		Alarm    : level,
		Message  : str,
		Filename : fileName,
		Line     : line,
		Funcname : functionName,
	}
	jsons, _ := json.Marshal(msg)

	if level == "EMAIL" {
		// 执行发邮件
		fmt.Println(jsons)

	} else if level == "SMS" {
		// 执行发短信
		fmt.Println(jsons)
	} else if level == "WX" {
		// 执行发微信
		fmt.Println(jsons)
	} else if level == "INFO" {
		// 执行记日志
		fmt.Println(jsons)
	} else if level == "PANIC" {
		// 执行PANIC方式
		fmt.Println(jsons)
	}
}
