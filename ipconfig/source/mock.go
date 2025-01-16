package source

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/discovery"
	"github.com/brelance/plato/common/logger"
)

func testServiceRegister(ctx *context.Context, port, node string) {
	// 模拟服务发现
	go func() {
		ed := discovery.EndpointInfo{
			IP:   "127.0.0.1",
			Port: port,
			MetaData: map[string]interface{}{
				"connect_num":   float64(rand.Int63n(12312321231231131)),
				"message_bytes": float64(rand.Int63n(1231232131556)),
			},
		}
		sr, err := discovery.NewServiceRegister(ctx, fmt.Sprintf("%s/%s", config.GetServicePathFromIPConf(), node), &ed, time.Now().Unix())
		logger.Logger.Info().Msg("register success")
		if err != nil {
			panic(err)
		}
		go sr.ListenLeaseRespChan()
		for {
			ed = discovery.EndpointInfo{
				IP:   "127.0.0.1",
				Port: port,
				MetaData: map[string]interface{}{
					"connect_num":   float64(rand.Int63n(12312321231231131)),
					"message_bytes": float64(rand.Int63n(1231232131556)),
				},
			}
			sr.UpdateValue(&ed)
			time.Sleep(1 * time.Second)
		}
	}()
}
