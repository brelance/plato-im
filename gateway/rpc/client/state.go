package client

import (
	"context"
	"time"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/prpc"
	service "github.com/brelance/plato/state/rpc/server"
)

var stateClient service.StateClient

func initStateClient() {
	// Why
	pCli, err := prpc.NewPClient(config.GetStateServiceName())
	if err != nil {
		panic(err)
	}
	// Why
	cli, err := pCli.DialByEndPoint(config.GetGatewayStateServerEndPoint())
	if err != nil {
		panic(err)
	}
	stateClient = service.NewStateClient(cli)
}

// RPC call
func CancelConn(ctx *context.Context, endpoint string, connID uint64, PayLoad []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	_, err := stateClient.CancelConn(rpcCtx, &service.StateRequest{
		Endpoint: endpoint,
		ConnID:   connID,
		Data:     PayLoad,
	})
	if err != nil {
		return err
	}
	return nil
}

func SendMsg(ctx *context.Context, endpoint string, connID uint64, PayLoad []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	_, err := stateClient.SendMsg(rpcCtx, &service.StateRequest{
		Endpoint: endpoint,
		ConnID:   connID,
		Data:     PayLoad,
	})
	if err != nil {
		return err
	}
	return nil
}
