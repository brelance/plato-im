package config

import "github.com/spf13/viper"

func GetEpollWaitQueueSize() uint {
	return viper.GetUint("gateway.epoll_wait_queue_size")
}

func GetGatewayEpollerNum() uint {
	return viper.GetUint("gateway.epoll_num")
}

func GetMaxTCPNum() int32 {
	return viper.GetInt32("gateway.epoll_num")

}
