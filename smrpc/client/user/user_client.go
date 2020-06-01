package user

import (
	"context"
	"github.com/smallnest/rpcx/client"
	"newrpc/domain/user"
	"newrpc/rpc"
)

var userRpc *UserRPC

func GetUserRPC() *UserRPC {
	if userRpc == nil {
		return newUserRPC()
	} else {
		return userRpc
	}
}
type UserRPC struct {
	Client client.XClient
}
func newUserRPC() *UserRPC {
	name := "UserRPC"
	userRpc = &UserRPC{
		Client: rpc.NewClient(name),
	}
	return userRpc
}

func (r *UserRPC) Regist(ctx context.Context, user_in *user.User_in, user_out *user.User_out) error {
	return r.Client.Call(ctx, "Regist", user_in, user_out)
}
