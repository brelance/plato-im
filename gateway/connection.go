package gateway

import (
	"net"
	"sync/atomic"
)

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

func newConnection(conn *net.TCPConn) *connection {
	return &connection{
		id:   atomic.AddUint64(&connId, 1),
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
