package service

import (
	"context"

	"github.com/brelance/plato/common/logger"
)

// RPC

const (
	CancelConnCmd = 1
	SendMsgCmd    = 2
)

type CmdContext struct {
	Ctx      *context.Context
	Cmd      int32
	Endpoint string
	ConnID   uint64
	Payload  []byte
}

type Service struct {
	CmdChannel chan *CmdContext
	UnimplementedStateServer
}

func (service *Service) CancelConn(ctx context.Context, sr *StateRequest) (*StateResponse, error) {
	c := context.TODO()
	cmd := &CmdContext{
		Ctx:      &c,
		Cmd:      CancelConnCmd,
		Endpoint: sr.GetEndpoint(),
		ConnID:   sr.GetConnID(),
		Payload:  sr.GetData(),
	}
	service.CmdChannel <- cmd
	logger.Logger.
		Debug().
		Msgf("Cancel connection: endpoint=%s, connID=%d, chanLen=%d", cmd.Endpoint, cmd.ConnID, len(service.CmdChannel))
	return &StateResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}
func (service *Service) SendMsg(ctx context.Context, sr *StateRequest) (*StateResponse, error) {
	c := context.TODO()
	cmd := &CmdContext{
		Ctx:      &c,
		Cmd:      SendMsgCmd,
		Endpoint: sr.GetEndpoint(),
		ConnID:   sr.GetConnID(),
		Payload:  sr.GetData(),
	}
	service.CmdChannel <- cmd
	logger.Logger.
		Debug().
		Msgf("Send message: endpoint=%s, connID=%d, chanLen=%d", cmd.Endpoint, cmd.ConnID, len(service.CmdChannel))
	return &StateResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}
