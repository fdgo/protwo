package main

import (
	"context"
	"flag"

	example "github.com/rpcxio/rpcx-examples"
	"github.com/smallnest/rpcx/server"
)

var (
	addr = flag.String("addr", "localhost:8972", "server address")
)

type Arith struct{}

// the second parameter is not a pointer
func (t *Arith) Mul(ctx context.Context, args example.Args, reply *example.Reply) error {
	reply.C = args.A * args.B*2
	return nil
}

func main() {
	flag.Parse()

	s := server.NewServer()
	configCORS(s)
	//s.Register(new(Arith), "")
	s.RegisterName("Arith", new(Arith), "")
	err := s.Serve("tcp", *addr)
	if err != nil {
		panic(err)
	}
}

func configCORS(s *server.Server) {
	opt := server.AllowAllCORSOptions()

	s.SetCORS(opt)
}
