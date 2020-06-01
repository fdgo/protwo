package main

import (
	"fmt"
	"github.com/wangmhgo/go-project/gdun/net/wss"
	"github.com/wangmhgo/go-project/gdun/server/booking"
)

func main() {
	_, err := booking.NewServer("", "192.168.60.10:6379,192.168.60.11:6379,192.168.60.12:6379,192.168.60.13:6379,192.168.60.14:6379,192.168.60.15:6379")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	sv := wss.NewWebSocketServer("", "", "", "htdoc", false, "log")
	sv.Start()
}
