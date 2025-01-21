package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/logger"
	"github.com/brelance/plato/common/prpc"
	"github.com/brelance/plato/common/tcp"
	"github.com/brelance/plato/gateway/rpc/client"
	service "github.com/brelance/plato/gateway/rpc/server"
	"google.golang.org/grpc"
)

var cmdChannel chan *service.CmdContext

func RunMain(path string) {
	config.Init(path)
	ln, err := net.ListenTCP("TCP", &net.TCPAddr{Port: config.GetGatewayTCPServerPort()})

	if err != nil {
		panic(err)
	}
	initWorkPoll()
	InitEpool(ln, runProc)

	fmt.Println("-------------im gateway stated------------")
	cmdChannel = make(chan *service.CmdContext, config.GetGatewayCmdChannelNum())
	s := prpc.NewPServer(
		prpc.WithServiceName(config.GetGatewayServiceName()),
		prpc.WithIP(config.GetGatewayServiceAddr()),
		prpc.WithPort(config.GetGatewayRPCServerPort()), prpc.WithWeight(config.GetGatewayRPCWeight()))
	fmt.Println(config.GetGatewayServiceName(), config.GetGatewayServiceAddr(), config.GetGatewayRPCServerPort(), config.GetGatewayRPCWeight())
	s.RegisterService(func(server *grpc.Server) {
		service.RegisterGatewayServer(server, &service.Service{CmdChannel: cmdChannel})
	})
	// 启动rpc 客户端
	client.Init()
	// 启动 命令处理写协程
	go cmdHandler()
	// 启动 rpc server
	s.Start(context.TODO())

}

func runProc(c *connection, ep *epoller) {
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

// state server(RPC client) -> gateway server(RPC server)
func cmdHandler() {
	for cmd := range cmdChannel {
		// 异步提交到协池中完成发送任务
		switch cmd.Cmd {
		case service.DelConnCmd:
			wPool.Submit(func() { closeConn(cmd) })
		case service.PushCmd:
			wPool.Submit(func() { sendMsgByCmd(cmd) })
		default:
			panic("command undefined")
		}
	}
}

// gateway client(RPC client) -> state server(RPC server)
func closeConn(cmd *service.CmdContext) {
	if connPtr, ok := epool.tables.Load(cmd.ConnID); ok {
		conn, _ := connPtr.(*connection)
		conn.Close()
	}
}
func sendMsgByCmd(cmd *service.CmdContext) {
	if connPtr, ok := epool.tables.Load(cmd.ConnID); ok {
		conn, _ := connPtr.(*connection)
		dp := tcp.DataPkg{
			Len:  uint32(len(cmd.Payload)),
			Data: cmd.Payload,
		}
		tcp.SendData(conn.conn, dp.Marshal())
	}
}

func getEndpoint() string {
	return fmt.Sprintf("%s:%d", config.GetGatewayServiceAddr(), config.GetGatewayRPCServerPort())
}
