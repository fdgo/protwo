package base

import (
	"context"
	"github.com/smallnest/rpcx/client"
	"newrpc/domain/base"
	"newrpc/rpc"
)

var baseRpc *BaseRPC

func GetBaseRpc() *BaseRPC {
	if baseRpc == nil {
		return newBaseRPC()
	} else {
		return baseRpc
	}
}
type BaseRPC struct {
	Client client.XClient
}
func newBaseRPC() *BaseRPC {
	name := "BaseRPC"
	baseRpc = &BaseRPC{
		Client: rpc.NewClient(name),
	}
	return baseRpc
}
func (r *BaseRPC) VfCode(ctx context.Context, base_in *base.Base_in, base_out *base.Base_out) error {
	return r.Client.Call(ctx, "VfCode", base_in, base_out)
}
