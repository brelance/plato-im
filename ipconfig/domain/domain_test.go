package domain

import (
	"testing"
	"time"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/logger"
	"github.com/brelance/plato/ipconfig/source"
)

func TestDomain(t *testing.T) {
	config.Init("/home/brelance/local/plato/plato.yaml")
	logger.Init()
	source.Init()
	Init()
	time.Sleep(1000 * time.Second)
}
