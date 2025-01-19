package gateway

import (
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
	// create sender of eChan
	epool.createAcceptProcess()
	//  create receiver of eChan
	epool.startEPool()
}

func (e *ePool) createAcceptProcess() {
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				c, err := e.ln.AcceptTCP()
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						logger.Logger.Error().Msgf("accept timeout: %v", ne)
					}
					logger.Logger.Error().Msgf("err: %v", err)
				}

				if !checkTCPLimit() {
					c.Close()
					continue
				}
				setTCPConfig(c)
				conn := newConnection(c)
				epool.addTask(conn)
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

// Start eSize epoller
func (e *ePool) startEPool() {
	for i := 0; i < e.eSize; i++ {
		e.startEProc()
	}
}

func (e *ePool) startEProc() {
	ep, err := newEpoller()
	if err != nil {
		panic(err)
	}
	// the receiver of eChan
	go func() {
		for {
			select {
			case <-e.done:
				return
			case conn := <-e.eChan:
				addTCPNum()
				logger.Logger.Debug().Msgf("tcpNum: %d", getTCPNum())
				if err := ep.add(conn); err != nil {
					logger.Logger.Error().Msgf("Failed to add connection %v\n", err)
					conn.Close()
					continue
				}
				logger.Logger.Debug().Msgf("EpollerPool new connection[%v] tcpSize:%d\n", conn.RemoteAddr(), tcpNum)
			}
		}
	}()

	for {
		select {
		case <-e.done:
			return
		default:
			connList, err := ep.wait(200)
			if err != nil && err != syscall.EINTR {
				logger.Logger.Error().Msgf("falied to epoll waith %v\n", err)
			}
			for _, conn := range connList {
				if conn == nil {
					break
				}
				epool.f(conn, ep)
			}
		}
	}
}

func (e *ePool) addTask(conn *connction) {
	e.eChan <- conn
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
	// bind epoller
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

// Allow gateway process to acquire the max number of resources, such as fd.
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

func setTCPConfig(conn *net.TCPConn) {
	conn.SetKeepAlive(true)
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
