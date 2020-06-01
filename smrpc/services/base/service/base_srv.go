package service

import (
	"context"
	"newrpc/domain/base"
)

type BaseRPC struct{}
func (*BaseRPC) VfCode(ctx context.Context, in *base.Base_in, out *base.Base_out) error {
	(*out).Name="vfcode 111"
	return nil
}
