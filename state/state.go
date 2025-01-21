package state

import (
	"fmt"
	"sync"
	"time"

	"context"

	"github.com/brelance/plato/common/cache"
	"github.com/brelance/plato/common/router"
	"github.com/brelance/plato/common/timingwheel"
	"github.com/brelance/plato/state/rpc/client"
)

type connState struct {
	// why we use RW lock here
	sync.RWMutex
	//
	heartTimer *timingwheel.Timer
	// runing after heartTimer event fired
	reConnTimer  *timingwheel.Timer
	msgTimer     *timingwheel.Timer
	msgTimerLock string
	connID       uint64
	did          uint64
}

func (c *connState) close(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()
	if c.heartTimer != nil {
		c.heartTimer.Stop()
	}
	if c.reConnTimer != nil {
		c.reConnTimer.Stop()
	}
	if c.msgTimer != nil {
		c.msgTimer.Stop()
	}
	loginSlotKey := cs.getLoginSlotKey(c.connID)
	meta := cs.loginSlotMarshal(c.did, c.connID)
	err := cache.SREM(ctx, loginSlotKey, meta)
	if err != nil {
		return err
	}
	// delete maxclientid
	slot := cs.getConnStateSlot(c.connID)
	key := fmt.Sprintf(cache.MaxClientIDKey, slot, c.connID, "*")
	keys, err := cache.GetKeys(ctx, key)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		err = cache.Del(ctx, keys...)
		if err != nil {
			return err
		}
	}
	// delete record in router table
	err = router.DelRecord(ctx, c.did)
	if err != nil {
		return err
	}
	// delete last message
	lastMsg := fmt.Sprintf(cache.LastMsgKey, slot, c.connID)
	err = cache.Del(ctx, lastMsg)
	if err != nil {
		return err
	}
	// send to gateway rpc server
	err = client.DelConn(&ctx, c.connID, nil)
	if err != nil {
		return err
	}
	// delete connID -> state in local
	cs.deleteConnIDState(ctx, c.connID)
	return nil
}

func (c *connState) appendMsg(ctx context.Context, key, msgTimerLock string, msgData []byte) {
	c.Lock()
	defer c.Unlock()
	c.msgTimerLock = msgTimerLock
	// reset msg timer. why
	if c.msgTimer != nil {
		c.msgTimer.Stop()
		c.msgTimer = nil
	}

	// set msg timer. If the process does not receive ack in 100ms, the process will resend the message.
	t := AfterFunc(100*time.Millisecond, func() {
		rePush(c.connID)
	})

	c.msgTimer = t
	err := cache.SetBytes(ctx, key, msgData, cache.TTL7D)
	if err != nil {
		panic(err)
	}
}

func (c *connState) loadMsgTimer(ctx context.Context) {
	data, err := cs.getLastMsg(ctx, c.connID)
	if err != nil {
		panic(err)
	}
	if data == nil {
		return
	}
	c.resetMsgTimer(c.connID, data.SessionID, data.MsgID)
}

func (c *connState) resetMsgTimer(connID, sessionID, msgID uint64) {
	c.Lock()
	defer c.Unlock()
	if c.msgTimer != nil {
		c.msgTimer.Stop()
	}
	c.msgTimerLock = fmt.Sprintf("%d_%d", sessionID, msgID)
	c.msgTimer = AfterFunc(100*time.Millisecond, func() {
		rePush(connID)
	})
}

func (c *connState) resetHeartbeatTimer() {
	c.Lock()
	defer c.Unlock()
	if c.heartTimer != nil {
		c.heartTimer.Stop()
	}

	c.heartTimer = AfterFunc(5*time.Second, func() {
		c.reSetReConnTimer()
	})
}

func (c *connState) reSetReConnTimer() {
	c.Lock()
	defer c.Unlock()

	if c.reConnTimer != nil {
		c.reConnTimer.Stop()
	}

	c.reConnTimer = AfterFunc(10*time.Second, func() {
		ctx := context.TODO()
		cs.connLogOut(ctx, c.connID)
	})
}

// handler ack from download client
func (c *connState) ackLastMsg(ctx context.Context, sessionID, msgID uint64) bool {
	c.Lock()
	defer c.Unlock()
	msgTimerLock := fmt.Sprintf("%d_%d", sessionID, msgID)
	if c.msgTimerLock != msgTimerLock {
		return false
	}
	slot := cs.getConnStateSlot(c.connID)
	key := fmt.Sprintf(cache.LastMsgKey, slot, c.connID)
	if err := cache.Del(ctx, key); err != nil {
		return false
	}
	if c.msgTimer != nil {
		c.msgTimer.Stop()
	}
	return true
}
