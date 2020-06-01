package user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	userdn "newrpc/domain/user"
	userclt "newrpc/client/user"
	"newrpc/support/utils/errex"
	"newrpc/support/utils/param"
	rsp "newrpc/support/utils/rsp"
)

func Regist(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	var in userdn.User_in
	if err := c.ShouldBindJSON(&in); err != nil {
		rsp.RespGin(400, int32(errex.NORMAL_INVALID_PARAMETER), "输入有误,请重写输入!", "参数有误", err.Error(), c)
		return
	}
	isok, _ := param.IsParam(in)
	if !isok {
		rsp.RespGin(400, int32(errex.NORMAL_INVALID_PARAMETER), "输入有误,请重写输入!", "参数有误", "参数有误", c)
		return
	}
	var out userdn.User_out
	user_rpc := userclt.GetUserRPC()
	err := user_rpc.Regist(c, &in, &out)
	if err != nil {
		fmt.Printf("Call Regist error:%s\n", err.Error())
		c.JSON(200, err)
		return
	}
	fmt.Println(out)

	c.JSON(200, out)
}

//****************************************************************************
//var upgrader = websocket.Upgrader{
//	CheckOrigin: func(r *http.Request) bool { return true },
//}
//
//func isExpectedClose(err error) bool {
//	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
//		log.Println("Unexpected websocket close: ", err)
//		return false
//	}
//	return true
//}
//func Do(cli useproto.UserService, ws *websocket.Conn) error {
//	var req useproto.WsIn
//	err := ws.ReadJSON(&req)
//	if err != nil {
//		return err
//	}
//	go func() {
//		for {
//			if _, _, err := ws.NextReader(); err != nil {
//				break
//			}
//		}
//	}()
//	log.Printf("Received Request: %v", req)
//	stream, err := cli.ServerStream(context.Background(), &req)
//	if err != nil {
//		return err
//	}
//	defer stream.Close()
//	for {
//		rsp, err := stream.Recv()
//		if err != nil {
//			if err != io.EOF {
//				return err
//			}
//			break
//		}
//		fmt.Println("888:", rsp.Data)
//		err = ws.WriteJSON(string(rsp.Data))
//		if err != nil {
//			if isExpectedClose(err) {
//				log.Println("Expected Close on socket", err)
//				break
//			} else {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//func WebsocketMsg(c *gin.Context) {
//	defer func() {
//		if err := recover(); err != nil {
//			fmt.Println(err)
//		}
//	}()
//	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		log.Fatal("Upgrade: ", err)
//		return
//	}
//	defer conn.Close()
//	if err := Do(client.UserClient, conn); err != nil {
//		log.Fatal("Echo: ", err)
//		return
//	}
//}
