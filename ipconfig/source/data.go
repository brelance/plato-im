package source

import (
	"context"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/discovery"
	"github.com/brelance/plato/common/logger"
)

func Init() {
	eventChan = make(chan *Event)
	ctx := context.Background()
	logger.Logger.Info().Msg("Init")
	go Datahandler(&ctx)
	if config.IsDebug() {
		ctx := context.Background()
		testServiceRegister(&ctx, "8896", "node1")
		testServiceRegister(&ctx, "8897", "node2")
		testServiceRegister(&ctx, "8898", "node3")
	}
	// debug
}

// core handler function
func Datahandler(ctx *context.Context) {
	dis := discovery.NewServiceDiscovery(ctx)
	defer dis.Close()

	setFunc := func(key string, value string) {
		if ed, err := discovery.UnMarshal([]byte(value)); err == nil {
			if event := NewEvent(ed); event != nil {
				event.Type = AddNodeEvent
				eventChan <- event
			}
		} else {
			logger.Logger.
				Error().
				Msgf("DataHandler.setFunc.err :%s", err.Error())
		}
	}

	delFunc := func(key string, value string) {
		if ed, err := discovery.UnMarshal([]byte(value)); err == nil {
			if event := NewEvent(ed); event != nil {
				event.Type = DelNodeEvent
				eventChan <- event
			}
		} else {
			logger.Logger.
				Error().
				Msgf("DataHandler.delFunc.err :%s", err.Error())
		}
	}

	err := dis.WatchService(config.GetServicePathFromIPConf(), setFunc, delFunc)
	if err != nil {
		panic(err)
	}
}

func EventChan() <-chan *Event {
	return eventChan
}
