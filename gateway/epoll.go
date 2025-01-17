package gateway

import (
	"fmt"
	"net"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/logger"
	"golang.org/x/sys/unix"
)

var epool *ePool
var tcpNum int32

type ePool struct {
	eChan  chan *connction
	tables sync.Map
	// the max number of epoller
	eSize int
	done  chan struct{}
	ln    *net.TCPListener
	f     func(c *connction, ep *epoller)
}

func InitEpool(ln *net.TCPListener, f func(c *connction, ep *epoller)) {
	setLimit()
	epool = NewEpool(ln, f)

}

func (e *ePool) createAcceptProcess() {
	for i := range runtime.NumCPU() {
		go func() {
			for {
				c, err := e.ln.Accept()
				if err != nil {
					fmt.Errorf("err: %v", err)
				}

				if !checkTCPLimit() {
					c.Close()
					continue
				}
			}
		}()
	}
}

func NewEpool(ln *net.TCPListener, f func(c *connction, ep *epoller)) *ePool {
	return &ePool{
		eChan: make(chan *connction),
		eSize: int(config.GetGatewayEpollerNum()),
		done:  make(chan struct{}),
		ln:    ln,
		f:     f,
	}
}

type epoller struct {
	// unix epoll fd
	fd            int
	fdToConnTable sync.Map
}

func newEpoller() (*epoller, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoller{
		fd: fd,
	}, nil
}

func (ep *epoller) add(conn *connction) error {
	fd := conn.fd
	err := unix.EpollCtl(
		ep.fd,
		syscall.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}

	ep.fdToConnTable.Store(fd, conn)
	epool.tables.Store(conn.id, conn)
	conn.BindEpoller(ep)
	return nil
}

func (ep *epoller) remove(conn *connction) error {
	subTCPNum()
	fd := conn.fd
	err := unix.EpollCtl(
		ep.fd,
		syscall.EPOLL_CTL_DEL,
		fd,
		nil)
	if err != nil {
		return err
	}

	ep.fdToConnTable.Delete(fd)
	epool.tables.Delete(conn.id)
	return nil
}

func (ep *epoller) wait(msec int) ([]*connction, error) {
	events := make([]unix.EpollEvent, config.GetEpollWaitQueueSize())
	n, err := unix.EpollWait(ep.fd, events, msec)
	if err != nil {
		return nil, err
	}

	conns := make([]*connction, n)
	for index, e := range events {
		if conn, ok := ep.fdToConnTable.Load(e.Fd); ok {
			conns[index] = conn.(*connction)
		}
	}
	return conns, nil
}

func socketFD(conn *net.TCPConn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(*conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pdf")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

func setLimit() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	logger.Logger.Info().Msgf("Set cur limit: %d", rLimit.Cur)
}

func checkTCPLimit() bool {
	tcpNum := getTCPNum()
	maxTCPNum := config.GetMaxTCPNum()
	return tcpNum <= maxTCPNum
}

func addTCPNum() {
	atomic.AddInt32(&tcpNum, 1)
}

func subTCPNum() {
	atomic.AddInt32(&tcpNum, -1)
}

func getTCPNum() int32 {
	return atomic.LoadInt32(&tcpNum)
}
