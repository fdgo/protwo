package service

import (
	"fmt"
	"context"
	baseclt "newrpc/client/base"
	"newrpc/domain/user"
	"newrpc/domain/base"
)

type UserRPC struct{}

func (*UserRPC) Regist(ctx context.Context, in *user.User_in, out *user.User_out) error {
	var basein base.Base_in
	var baseout base.Base_out
	basein.Name = in.Name
	base_rpc := baseclt.GetBaseRpc()
	err := base_rpc.VfCode(ctx, &basein, &baseout)
	if err != nil {
		fmt.Printf("Call Regist error:%s\n", err.Error())
		return err
	}
	(*out).Name = baseout.Name
	return nil
}
