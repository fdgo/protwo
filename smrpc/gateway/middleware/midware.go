package middleware

import (
	token "newrpc/support/utils/auth"
	"newrpc/support/utils/cors"
	"newrpc/support/utils/errex"
	"newrpc/support/utils/logex"
	rsp "newrpc/support/utils/rsp"
	time_ex "newrpc/support/utils/timex"
	"newrpc/support/utils/trace"
	"github.com/gin-gonic/gin"
)

func Log() gin.HandlerFunc {
	return logex.GinLogger()
}

func TracerWrapper(c *gin.Context) {
	trace.TracerWrapper(c)
}

func Cors() gin.HandlerFunc {
	return cors.Cors()
}

func NoRoute() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == "OPTIONS" {
			ctx.JSON(200, nil)
		}
	}
}

func Auth(token *token.JwtToken) gin.HandlerFunc {
	return func(c *gin.Context) {
		istokenok, msg, sub := token.Decode(c.Request, c.Writer)
		if !istokenok {
			rsp.RespGin(400, int32(errex.NORMAL_NORMALUNAUTHORIZED), " 请先登录!", msg, "nil", c)
			c.Abort()
			return
		}
		c.Request.Header.Set("X-Head-Uuid", sub.Uuid)
		c.Request.Header.Set("X-Head-Mobile", sub.Mobile)
		c.Request.Header.Set("X-Head-UserName", sub.UserName)
		c.Request.Header.Set("X-Head-InvCodeAgent", sub.InvCodeAgent)
		c.Request.Header.Set("X-Head-InvCodeSelf", sub.InvCodeSelf)
		c.Request.Header.Set("X-Head-TimeStamp", time_ex.TimeStampToTimeStr(sub.ExpiresAt))
		c.Next()
	}
}