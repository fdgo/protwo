package main

import (
	"github.com/gin-gonic/gin"
	midware "newrpc/gateway/middleware"
	"newrpc/gateway/service/base"
	"newrpc/gateway/service/user"
)

func main() {
	gin.SetMode(gin.DebugMode)
	router := gin.New()
	router.Use(gin.Recovery())
	//router.Use(midware.Log())
	router.Use(midware.TracerWrapper)
	router.Use(midware.Cors())
	router.Use(midware.NoRoute())
	rf := router.Group("/api/v1")
	ru := rf.Group("/user")
	{
		ru.POST("/regist", user.Regist)
	}
	rb := rf.Group("/base")
	{
		rb.POST("/vfcode", base.VfCode)
	}
	router.Run() // listen and serve on 0.0.0.0:8080
}
