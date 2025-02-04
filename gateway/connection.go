package gateway

import (
	"errors"
	"net"
	"sync"
	"time"
)

const (
	version      = uint64(0)
	sequenceBits = uint64(16)
	maxSequence  = int64(-1) ^ (int64(-1) << int64(sequenceBits))
	timeLeft     = uint8(16)
	versionLeft  = uint8(63)
	twepoch      = int64(1589923200000)
)

var node *ConnIDGenerater

type ConnIDGenerater struct {
	mu        sync.Mutex
	LastStamp int64
	Sequence  int64
}

type connection struct {
	id   uint64
	fd   int
	ep   *epoller
	conn *net.TCPConn
}

var connId uint64

func (c *connection) BindEpoller(ep *epoller) {
	c.ep = ep
}

func NewConnection(conn *net.TCPConn) *connection {
	id, err := node.NextID()
	if err != nil {
		panic(err)
	}

	return &connection{
		id:   id,
		fd:   socketFD(conn),
		conn: conn,
	}
}

func (c *connection) Close() {
	epool.tables.Delete(c.id)

	// Here is the reason why we should use sync map
	if c.ep != nil {
		c.ep.fdToConnTable.Delete(c.fd)
	}
	err := c.conn.Close()
	panic(err)
}

func (c *connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *ConnIDGenerater) getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}

func (w *ConnIDGenerater) NextID() (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.nextID()
}

func (w *ConnIDGenerater) nextID() (uint64, error) {
	timeStamp := w.getMilliSeconds()
	if timeStamp < w.LastStamp {
		return 0, errors.New("time is moving backwards, waiting until")
	}

	if w.LastStamp == timeStamp {
		w.Sequence = (w.Sequence + 1) & maxSequence
		if w.Sequence == 0 {
			for timeStamp <= w.LastStamp {
				timeStamp = w.getMilliSeconds()
			}
		}
	} else {
		w.Sequence = 0
	}
	w.LastStamp = timeStamp
	id := ((timeStamp - twepoch) << int64(timeLeft)) | w.Sequence
	connID := uint64(id) | (version << uint64(versionLeft))
	return connID, nil
}
