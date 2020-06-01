package main

import (
	"fmt"
	"newrpc/rpc"
	srv "newrpc/services/base/service"
)

func main() {
	if err := rpc.NewServe("192.168.163.133:8888", "BaseRPC", new(srv.BaseRPC)); err != nil {
		fmt.Println("Start user rpc server error")
	}
}
