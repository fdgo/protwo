package rpc

import (
	"context"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"net"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

var (
	BasePath = "bifund-rpcx"
)

type TracePlugin struct {
	start time.Time
}

func IsPanicError(err error) bool {
	if nil == err {
		return false
	}
	// duplicate key
	if strings.Contains(err.Error(), "panic") ||
		strings.Contains(err.Error(), "runtime error") {
		return true
	}
	return false
}

func (p *TracePlugin) PostWriteResponse(ctx context.Context, req *protocol.Message, res *protocol.Message, err error) error {
	if err == nil {
		fmt.Println(fmt.Sprintf("CALL OK %s.%s seq:%d hand_time:%v", req.ServicePath, req.ServiceMethod, req.Seq(), time.Since(p.start)))
	} else {
		if IsPanicError(err) {
			fmt.Println(fmt.Sprintf("CALL FAILED %s.%s seq:%d hand_time:%v error:%s,stack:%s", req.ServicePath, req.ServiceMethod, req.Seq(), time.Since(p.start), err.Error(), string(debug.Stack())))
		} else {
			fmt.Println(fmt.Sprintf("CALL FAILED %s.%s seq:%d hand_time:%v error:%s", req.ServicePath, req.ServiceMethod, req.Seq(), time.Since(p.start), err.Error()))
		}
	}
	return nil

}

func (p *TracePlugin) Register(name string, rcvr interface{}, metadata string) error {
	fmt.Println(fmt.Sprintf("RPC Register %s", name))
	return nil
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (p *TracePlugin) HandleConnAccept(conn net.Conn) (net.Conn, bool) {
	fmt.Println(fmt.Sprintf("RPC accept client %s", conn.RemoteAddr().String()))
	return conn, true
}

func (p *TracePlugin) PostReadRequest(ctx context.Context, r *protocol.Message, e error) error {
	p.start = time.Now().UTC()
	return nil
}
func NewServe(serverAddr, name string, rcvr interface{}) error {
	s := server.NewServer()

	traceP := TracePlugin{}
	s.Plugins.Add(&traceP)

	//blacklist := &serverplugin.BlacklistPlugin{
	//	Blacklist: map[string]bool{"127.0.0.1": true},
	//}
	//s.Plugins.Add(blacklist)

	r := &serverplugin.RedisRegisterPlugin{
		ServiceAddress: "tcp@" + serverAddr,
		RedisServers:   []string{"120.27.239.127:6379"},
		BasePath:       BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Second * 10,
	}
	if err := r.Start(); err != nil {
		return err
	}
	s.Plugins.Add(r)
	s.RegisterName(name, rcvr, "")
	addr := strings.Split(serverAddr, ":")
	listenAddr := fmt.Sprintf(":%s", addr[1])
	if err := s.Serve("tcp", listenAddr); err != nil {
		return err
	}
	return nil
}
func NewClient(name string) client.XClient {
	d := client.NewRedisDiscovery(BasePath, name, []string{"120.27.239.127:6379"}, nil)
	services := d.GetServices()
	for i := 0; i < len(services); i++ {
		s := services[i]
		fmt.Printf("%s:%s", s.Key, s.Value)
	}
	option := client.DefaultOption
	option.SerializeType = protocol.JSON
	option.Heartbeat = true
	option.HeartbeatInterval = time.Second * 3
	xclient := client.NewXClient(name, client.Failtry, client.RandomSelect, d, option)
	return xclient
}
//wrk -t2 -c100 -d10s -s /opt/workspace/src/microservice/jzapi/post  --latency  "http://127.0.0.1:9908/user/login_account"
