package gateway

import (
	"github.com/brelance/plato/common/config"
	"github.com/panjf2000/ants"
)

var wPool *ants.Pool

func initWorkPoll() {
	var err error
	workPoolNum := config.GetGatewayEpollerNum()
	if wPool, err = ants.NewPool(int(workPoolNum)); err != nil {
		// logger.Logger.Error().Msgf("InitWorkPoll.err: %s nun:%d\n", err.Error(), workPoolNum)
		panic(err)
	}
}
