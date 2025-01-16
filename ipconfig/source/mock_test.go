package source

import (
	"testing"
	"time"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/logger"
)

func TestSource(t *testing.T) {
	config.Init("/home/brelance/local/plato/plato.yaml")
	logger.Init()
	println(config.IsDebug())
	Init()
	time.Sleep(1000 * time.Second)
}
