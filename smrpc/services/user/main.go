package main

import (
	"fmt"
	"newrpc/rpc"
	srv "newrpc/services/user/service"
)

func main() {
	if err := rpc.NewServe("192.168.163.133:9999", "UserRPC", new(srv.UserRPC)); err != nil {
		fmt.Println("Start user rpc server error")
	}
}
