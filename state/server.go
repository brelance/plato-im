package state

import (
	"context"

	"github.com/brelance/plato/common/idl/message"
	"github.com/brelance/plato/common/logger"
	"github.com/brelance/plato/state/rpc/client"
	service "github.com/brelance/plato/state/rpc/server"
	"google.golang.org/protobuf/proto"
)

// state RPC server
func cmdHandler() {
	for cmdCtx := range cs.server.CmdChannel {
		switch cmdCtx.Cmd {
		case service.CancelConnCmd:
			cs.connLogOut(*cmdCtx.Ctx, cmdCtx.ConnID)
		case service.SendMsgCmd:
			msgCmd := &message.MsgCmd{}
			err := proto.Unmarshal(cmdCtx.Payload, msgCmd)
			if err != nil {
				logger.Logger.
					Error().
					Msgf("Protobuf unmarshal error: %v", err)
			}
			msgCmdhandler(cmdCtx, msgCmd)
		}
	}
}

func msgCmdhandler(cmdCtx *service.CmdContext, msgCmd *message.MsgCmd) {
	switch msgCmd.Type {
	case message.CmdType_Login:
		loginMsgHandler(cmdCtx, msgCmd)
	case message.CmdType_Heartbeat:
		heartbeatMsgHandler(cmdCtx, msgCmd)
	case message.CmdType_ReConn:
		reConnMsgHandler(cmdCtx, msgCmd)
	case message.CmdType_UP:
		upMsgHandler(cmdCtx, msgCmd)
	case message.CmdType_ACK:
		ackMsgHandler(cmdCtx, msgCmd)
	}
}

func loginMsgHandler(cmdCtx *service.CmdContext, msgCmd *message.MsgCmd) {
	loginMsg := &message.LoginMsg{}
	err := proto.Unmarshal(msgCmd.Payload, loginMsg)
	if err != nil {
		logger.Logger.
			Error().
			Msgf("protobuf unmarshal loginmsg error: %s", err)
	}

	if loginMsg.Head != nil {
		logger.Logger.
			Debug().
			Msg("loginMsgHandler")
	}

	err = cs.connLogin(*cmdCtx.Ctx, loginMsg.Head.DeviceID, cmdCtx.ConnID)
	if err != nil {
		logger.Logger.
			Error().
			Msgf("cache state login error: %s", err)
	}
	sendACKMsg(message.CmdType_Login, cmdCtx.ConnID, 0, 0, "login success")
}

// done heartbeat timer 5s -> reconn timer 10s
func heartbeatMsgHandler(cmdCtx *service.CmdContext, msgCmd *message.MsgCmd) {
	heartbeatMsg := &message.HeartbeatMsg{}
	err := proto.Unmarshal(msgCmd.Payload, heartbeatMsg)
	if err != nil {
		logger.Logger.
			Error().
			Msgf("protobuf unmarshal heartbeat error: %s", err)
	}
	cs.resetHeartbeatTimer(cmdCtx.ConnID)
}

func reConnMsgHandler(cmdCtx *service.CmdContext, msgCmd *message.MsgCmd) {
	reConnMsg := &message.ReConnMsg{}
	err := proto.Unmarshal(msgCmd.Payload, reConnMsg)
	if err != nil {
		logger.Logger.
			Error().
			Msgf("protobuf unmarshal reConnmsg error: %s", err)
	}
	// TODO
	var (
		code uint32
		msg  string
	)
	err = cs.reConn(*cmdCtx.Ctx, reConnMsg.Head.ConnID, cmdCtx.ConnID)
	if err != nil {
		code, msg = 1, "reconn falied"
		logger.Logger.
			Error().
			Msgf("reconnect error: %s", err)
	}
	sendACKMsg(message.CmdType_ReConn, cmdCtx.ConnID, 0, code, msg)
}

// hanlder the message send by up client
func upMsgHandler(cmdCtx *service.CmdContext, msgCmd *message.MsgCmd) {
	upMsg := &message.UPMsg{}
	err := proto.Unmarshal(msgCmd.Payload, upMsg)
	if err != nil {
		logger.Logger.
			Error().
			Msgf("protobuf unmarshal UPmsg error: %s", err)
	}

	// push the message to download cilent/logic level
	if cs.compareAndIncrClientID(*cmdCtx.Ctx, cmdCtx.ConnID, upMsg.Head.ClientID, upMsg.Head.SessionId) {
		sendACKMsg(message.CmdType_UP, cmdCtx.ConnID, 0, 0, "upload message success")
		// for testing TODO: we should call logic layer here to handler download message
		pushMsg(*cmdCtx.Ctx, cmdCtx.ConnID, cs.msgID, 0, upMsg.UPMsgBody)
	}
}

// handler the ack send by download client
func ackMsgHandler(cmdCtx *service.CmdContext, msgCmd *message.MsgCmd) {
	ackMsg := &message.ACKMsg{}
	err := proto.Unmarshal(msgCmd.Payload, ackMsg)
	if err != nil {
		logger.Logger.
			Error().
			Msgf("protobuf unmarshal ACKmsg error: %s", err)
	}
	// the connID in ackMsg is upload client's connID
	cs.ackLastMsg(*cmdCtx.Ctx, ackMsg.ConnID, ackMsg.SessionID, ackMsg.MsgID)
}

func sendACKMsg(ackType message.CmdType, connID, clientID uint64, code uint32, msg string) {
	ackMsg := &message.ACKMsg{}
	ackMsg.Type = ackType
	ackMsg.Code = code
	ackMsg.Msg = msg
	ackMsg.ConnID = connID
	ackMsg.ClientID = clientID
	download, err := proto.Marshal(ackMsg)
	if err != nil {
		logger.Logger.
			Error().
			Msgf("err: %s", err)
	}
	sendMsg(connID, message.CmdType_ACK, download)
}

func sendMsg(connID uint64, ty message.CmdType, downLoad []byte) {
	mc := &message.MsgCmd{}
	mc.Type = ty
	mc.Payload = downLoad
	data, err := proto.Marshal(mc)
	if err != nil {
		logger.Logger.
			Error().
			Msgf("SendMsg error: %s", err)
	}
	ctx := context.TODO()
	client.Push(&ctx, connID, data)
}

func pushMsg(ctx context.Context, connID, sessionID, msgID uint64, data []byte) {
	pushMsg := &message.PushMsg{
		Content: data,
		MsgID:   cs.msgID,
	}

	if data, err := proto.Marshal(&message.PushMsg{}); err != nil {
		logger.Logger.
			Error().
			Msgf("protobuf marshal push message error: %s", err)
	} else {
		sendMsg(connID, message.CmdType_Push, data)
		err = cs.appendLstMsg(ctx, connID, pushMsg)
		if err != nil {
			panic(err)
		}
	}
}

// use for download user
// situation: server -> download user. If the push message lost, the client will resend the message to download user.
func rePush(connID uint64) {
	pushMsg, err := cs.getLastMsg(context.Background(), connID)
	if err != nil {
		panic(err)
	}
	if pushMsg == nil {
		return
	}
	msgData, err := proto.Marshal(pushMsg)
	if err != nil {
		panic(err)
	}
	sendMsg(connID, message.CmdType_Push, msgData)
	if state, ok := cs.loadConnIDState(connID); ok {
		state.resetMsgTimer(connID, pushMsg.SessionID, pushMsg.MsgID)
	}
}
