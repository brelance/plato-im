package client

import (
	"context"
	"time"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/logger"
	"github.com/brelance/plato/common/prpc"
	service "github.com/brelance/plato/gateway/rpc/server"
)

var gatewayClient service.GatewayClient

func initGatewayClient() {
	pCli, err := prpc.NewPClient(config.GetGatewayServiceName())
	if err != nil {
		panic(err)
	}
	conn, err := pCli.DialByEndPoint(config.GetStateServerGatewayServerEndpoint())
	if err != nil {
		panic(err)
	}
	gatewayClient = service.NewGatewayClient(conn)
}

func DelConn(ctx *context.Context, connID uint64, Payload []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	gatewayClient.DelConn(rpcCtx, &service.GatewayRequest{ConnID: connID, Data: Payload})
	return nil
}

func Push(ctx *context.Context, connID uint64, Payload []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	resp, err := gatewayClient.Push(rpcCtx, &service.GatewayRequest{ConnID: connID, Data: Payload})
	if err != nil {
		logger.Logger.
			Error().
			Msgf("push error: %v", err)
	}
	logger.Logger.
		Debug().
		Msgf("%v", resp)
	return nil
}
