package gateway

import "net"

type connction struct {
	id   uint64
	fd   int
	e    *epoller
	conn *net.TCPConn
}

func (c *connction) BindEpoller(ep *epoller) {
	c.e = ep
}

func newConnection(conn *net.TCPConn) *connction {
	return &connction{
		id: ,
		fd: socket,
		conn: conn,
	}
}