package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/logger"
	"github.com/brelance/plato/common/tcp"
	"github.com/brelance/plato/gateway/rpc/client"
)

func RunMain(path string) {
	config.Init(path)
	ln, err := net.ListenTCP("TCP", &net.TCPAddr{Port: config.GetGatewayTCPServerPort()})

	if err != nil {
		logger.Logger.Error().Msgf("err: %v", err)
		os.exit(1)
	}

}

func runProc(c *connction, ep *epoller) {
	ctx := context.Background()
	dataBuf, err := tcp.ReadData(c.conn)
	if err != nil {
		// if the error is the end of the file.
		if errors.Is(err, io.EOF) {
			ep.remove(c)
			// why
			client.CancelConn(&ctx, getEndpoint(), c.id, nil)
		}
		return
	}

	err = wPool.Submit(func() {
		client.SendMsg(&ctx, getEndpoint(), c.id, dataBuf)
	})
	if err != nil {
		logger.Logger.Error().Msgf("err: %v", err)
	}
}

func getEndpoint() string {
	return fmt.Sprintf("%s:%d", config.GetGatewayServiceAddr(), config.GetGatewayRPCServerPort())
}
