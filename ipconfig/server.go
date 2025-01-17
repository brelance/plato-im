package ipconfig

import (
	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/ipconfig/domain"
	"github.com/brelance/plato/ipconfig/source"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func RunMain(path string) {
	config.Init(path)
	source.Init()
	domain.Init()
	s := server.Default(server.WithHostPorts(":6789"))
	s.GET("/ip/list", GetIpconfList)
	s.Spin()
}
