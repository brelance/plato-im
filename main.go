package main

import (
	"context"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/logger"
)

func main() {
	config.Init("/home/brelance/local/plato/plato.yaml")
	logger.Init()
	ctx := context.Background()
	logger.Logger.Info().Msg("this is output from main")

	done := make(chan bool)
	go func() {
		child(&ctx)
		done <- true
	}()
	<-done
}

func child(ctx *context.Context) {
	logger.Logger.Info().Msg("this is output from go routine")
}
