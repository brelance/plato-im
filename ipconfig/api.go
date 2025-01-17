package ipconfig

import (
	"context"

	"github.com/brelance/plato/ipconfig/domain"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type Response struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

func GetIpconfList(ctx context.Context, appCtx *app.RequestContext) {
	defer func() {
		if err := recover(); err != nil {
			appCtx.JSON(consts.StatusBadRequest, utils.H{"err": err})
		}
	}()

	ipConfCtx := domain.NewIpConfContext(&ctx, appCtx)
	eds := domain.Dispatch(ipConfCtx)
	ipConfCtx.AppCtx.JSON(consts.StatusOK, packRes(top5Endports(eds)))

}
