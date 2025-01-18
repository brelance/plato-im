package gateway

import (
	"net"
	"sync/atomic"
)

type connction struct {
	id   uint64
	fd   int
	ep   *epoller
	conn *net.TCPConn
}

var connId uint64

func (c *connction) BindEpoller(ep *epoller) {
	c.ep = ep
}

func newConnection(conn *net.TCPConn) *connction {
	return &connction{
		id:   atomic.AddUint64(&connId, 1),
		fd:   socketFD(conn),
		conn: conn,
	}
}

func (c *connction) Close() {
	epool.tables.Delete(c.id)

	// Here is the reason why we should use sync map
	if c.ep != nil {
		c.ep.fdToConnTable.Delete(c.fd)
	}
	err := c.conn.Close()
	panic(err)
}

func (c *connction) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}
