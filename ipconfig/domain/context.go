package domain

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

type IpConfContext struct {
	Ctx       *context.Context
	AppCtx    *app.RequestContext
	ClientCtx *ClientContext
}

type ClientContext struct {
	IP string `json:IP`
}

func NewIpConfContext(ctx *context.Context, appCtx *app.RequestContext) *IpConfContext {
	ipConfContext := &IpConfContext{
		Ctx:       ctx,
		AppCtx:    appCtx,
		ClientCtx: &ClientContext{},
	}

	return ipConfContext
}
