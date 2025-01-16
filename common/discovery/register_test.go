package dicovery

import (
	"context"
	"testing"
	"time"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/logger"
)

func TestServieRegister(t *testing.T) {
	ctx := context.Background()
	config.Init("/home/brelance/local/plato/plato.yaml")

	logger.Init()
	ser, err := NewServiceRegister(&ctx, "web/node1", &EndpointInfo{
		IP:   "127.0.0.1",
		Port: "9999",
	}, 5)

	if err != nil {
		panic(err)
	}

	go ser.ListenLeaseRespChan()
	select {
	case <-time.After(20 * time.Second):
		ser.Close()
	}
}
