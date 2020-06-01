package limit

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"newrpc/support/utils/timex"
	"runtime/debug"
	"strings"
	"time"
)

var (
	rl = New(2, time.Second)
)

// 中间件，用令牌桶限制请求频率
func LimitHandler(c *gin.Context) {
	if rl.Limit() {

		c.JSON(
			http.StatusOK,
			gin.H{
				"code":    9999,
				"message": "请求频率太高",
			},
		)
		c.Abort()
		return
	}
	c.Next()
}
func Recover(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				DebugStack := ""
				for _, v := range strings.Split(string(debug.Stack()), "\n") {
					DebugStack += v + "<br>"
				}
				str := name + timex.GetCurrentTime() + "|" + c.Request.Host + "|" + c.Request.RequestURI + "|" + c.Request.Method + "|" + DebugStack + "|" + c.Request.UserAgent()
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "系统异常，请联系管理员！",
					"err": str,
				})
			}
		}()
		c.Next()
	}
}
