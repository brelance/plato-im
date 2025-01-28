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

func GetGatewayWorkPoolNum() int {
	return viper.GetInt("gateway.work_pool_num")
}

func GetGatewayServiceName() string {
	return viper.GetString("gateway.service_name")
}

func GetGatewayServiceAddr() string {
	return viper.GetString("gateway.service_addr")
}

func GetGatewayRPCServerPort() int {
	return viper.GetInt("gateway.rpc_server_port")
}

func GetGatewayTCPServerPort() int {
	return viper.GetInt("gateway.tcp_server_port")
}

func GetGatewayStateServerEndPoint() string {
	return viper.GetString("gateway.state_server_endpoint")
}

func GetGatewayCmdChannelNum() int {
	return viper.GetInt("gateway.cmd_channel_num")
}

func GetGatewayRPCWeight() int {
	return viper.GetInt("gateway.weight")
}
