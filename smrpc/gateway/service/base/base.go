package base

import (
	"fmt"
	"github.com/gin-gonic/gin"
	baseclt "newrpc/client/base"
	basedn "newrpc/domain/base"
	"newrpc/support/utils/errex"
	"newrpc/support/utils/param"
	"newrpc/support/utils/rsp"
)

func VfCode(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	var in basedn.Base_in
	if err := c.ShouldBindJSON(&in); err != nil {
		rsp.RespGin(400, int32(errex.NORMAL_INVALID_PARAMETER), "输入有误,请重写输入!", "参数有误", err.Error(), c)
		return
	}
	isok, _ := param.IsParam(in)
	if !isok {
		rsp.RespGin(400, int32(errex.NORMAL_INVALID_PARAMETER), "输入有误,请重写输入!", "参数有误", "参数有误", c)
		return
	}
	var out basedn.Base_out
	base_rpc := baseclt.GetBaseRpc()
	err := base_rpc.VfCode(c, &in, &out)
	if err != nil {
		fmt.Printf("Call Regist error:%s\n", err.Error())
		c.JSON(200, err)
		return
	}
	fmt.Println(out)

	c.JSON(200, out)
}
