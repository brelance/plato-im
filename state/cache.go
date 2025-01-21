package state

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/brelance/plato/common/cache"
	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/idl/message"
	"github.com/brelance/plato/common/logger"
	"github.com/brelance/plato/common/router"
	service "github.com/brelance/plato/state/rpc/server"
	"google.golang.org/protobuf/proto"
)

var cs *cacheState

type cacheState struct {
	msgID            uint64
	connToStateTable sync.Map
	server           *service.Service
}

func InitCacheState(ctx context.Context) {
	cs := &cacheState{}
	cache.InitRedis(ctx)
	router.Init(ctx)
	cs.connToStateTable = sync.Map{}
	cs.server = &service.Service{
		CmdChannel: make(chan *service.CmdContext, config.GetSateCmdChannelNum()),
	}
	cs.initLoginSlot(ctx)
}

func (cs *cacheState) initLoginSlot(ctx context.Context) error {
	loginSlotRange := config.GetStateServerLoginSlotRange()
	for _, slot := range loginSlotRange {
		loginSlotKey := fmt.Sprintf(cache.LoginSlotSetKey, slot)
		go func() {
			loginSlot, err := cache.SmembersStrSlice(ctx, loginSlotKey)
			if err != nil {
				panic(err)
			}

			for _, meta := range loginSlot {
				did, connID := cs.loginSlotUnmarshal(meta)
				cs.connReLogin(ctx, did, connID)
			}

		}()
	}
	return nil
}

// clientID connID sessionID_msgID = seqID
func (cs *cacheState) connLogin(ctx context.Context, did, connID uint64) error {
	state := cs.newConnState(did, connID)
	loginSlotKey := cs.getLoginSlotKey(connID)
	meta := cs.loginSlotMarshal(did, connID)
	// Add login meta data to login slot stored in cache
	err := cache.SADD(ctx, loginSlotKey, meta)

	if err != nil {
		return err
	}

	// Add gateway address to router table.
	// why
	endPoint := fmt.Sprintf("%s:%d", config.GetGatewayServiceAddr(), config.GetStateServerPort())
	err = router.AddRecord(ctx, did, endPoint, connID)
	if err != nil {
		return err
	}

	cs.storeConnIDState(connID, state)
	return nil
}

func (cs *cacheState) connLogOut(ctx context.Context, connID uint64) (uint64, error) {
	if state, ok := cs.loadConnIDState(connID); ok {
		did := state.did
		return did, state.close(ctx)
	}
	return 0, nil
}

// How to use
func (cs *cacheState) connReLogin(ctx context.Context, did, connID uint64) {
	state := cs.newConnState(did, connID)
	cs.storeConnIDState(connID, state)
	state.loadMsgTimer(ctx)
}

func (cs *cacheState) loadConnIDState(connID uint64) (*connState, bool) {
	if data, ok := cs.connToStateTable.Load(connID); ok {
		state := data.(*connState)
		return state, true
	}
	return nil, false
}

func (cs *cacheState) reConn(ctx context.Context, oldConnID, newConnID uint64) error {
	var (
		did uint64
		err error
	)

	did, err = cs.connLogOut(ctx, oldConnID)
	if err != nil {
		return err
	}

	return cs.connLogin(ctx, did, newConnID)
}

// session level
func (cs *cacheState) compareAndIncrClientID(ctx context.Context, connID, oldMaxClientID uint64, sessionID string) bool {
	slot := cs.getConnStateSlot(connID)
	key := fmt.Sprintf(cache.MaxClientIDKey, slot, connID, sessionID)
	logger.Logger.
		Debug().
		Msgf("RunLuaInt %s, %d, %d\n", key, oldMaxClientID, cache.TTL7D)

	var (
		res int
		err error
	)

	if res, err = cache.RunLuaInt(ctx, cache.LuaCompareAndIncrClientID, []string{key}, oldMaxClientID, cache.TTL7D); err != nil {
		panic(err)
	}
	return res > 0
}

func (cs *cacheState) getLastMsg(ctx context.Context, connID uint64) (*message.PushMsg, error) {
	slot := cs.getConnStateSlot(connID)
	key := fmt.Sprintf(cache.LastMsgKey, slot, connID)
	data, err := cache.GetBytes(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	pushMsg := &message.PushMsg{}
	err = proto.Unmarshal(data, pushMsg)
	if err != nil {
		return nil, err
	}
	logger.Logger.
		Debug().
		Msgf("protobuf unmarshal last message: %v", pushMsg)

	return pushMsg, nil
}

func (cs *cacheState) appendLstMsg(ctx context.Context, connID uint64, pushMsg *message.PushMsg) error {
	if pushMsg == nil {
		logger.Logger.
			Error().
			Msgf("push message is empty")
	}

	var (
		state *connState
		ok    bool
	)
	if state, ok = cs.loadConnIDState(connID); !ok {
		return errors.New("connID state is nil")
	}
	slot := cs.getConnStateSlot(connID)
	key := fmt.Sprintf(cache.LastMsgKey, slot, connID)
	msgTimerLock := fmt.Sprintf("%d_%d", pushMsg.SessionID, pushMsg.MsgID)
	msgData, _ := proto.Marshal(pushMsg)
	state.appendMsg(ctx, key, msgTimerLock, msgData)

	return nil
}

func (cs *cacheState) ackLastMsg(ctx context.Context, connID, sessionID, msgID uint64) {
	var (
		state *connState
		ok    bool
	)
	if state, ok = cs.loadConnIDState(connID); ok {
		state.ackLastMsg(ctx, sessionID, msgID)
	}
}

func (cs *cacheState) resetHeartbeatTimer(connID uint64) {
	if data, ok := cs.connToStateTable.Load(connID); ok {
		state, _ := data.(*connState)
		state.resetHeartbeatTimer()
	}
}

func (cs *cacheState) newConnState(did, connID uint64) *connState {
	s := connState{
		connID: connID,
		did:    did,
	}
	s.resetHeartbeatTimer()
	return &s
}

func (cs *cacheState) storeConnIDState(connID uint64, state *connState) {
	cs.connToStateTable.Store(connID, state)
}

func (cs *cacheState) deleteConnIDState(ctx context.Context, connID uint64) {
	cs.connToStateTable.Delete(connID)
}

func (cs *cacheState) getConnStateSlot(connID uint64) uint64 {
	connStateSlotList := config.GetStateServerLoginSlotRange()
	slotSize := uint64(len(connStateSlotList))
	return connID % slotSize
}

func (cs *cacheState) getLoginSlotKey(connID uint64) string {
	connStateSlotList := config.GetStateServerLoginSlotRange()
	slotSize := uint64(len(connStateSlotList))
	slot := connID % slotSize
	slotKey := fmt.Sprintf(cache.LoginSlotSetKey, connStateSlotList[slot])
	return slotKey
}

func (cs *cacheState) loginSlotMarshal(did, connID uint64) string {
	return fmt.Sprintf("%d|%d", did, connID)
}

func (cs *cacheState) loginSlotUnmarshal(meta string) (uint64, uint64) {
	strs := strings.Split(meta, "|")
	if len(strs) < 2 {
		return 0, 0
	}
	did, err := strconv.ParseUint(strs[0], 10, 64)
	if err != nil {
		panic(err)
	}
	connID, err := strconv.ParseUint(strs[1], 10, 64)
	if err != nil {
		panic(err)
	}
	return did, connID
}
