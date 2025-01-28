package config

import (
	"time"

	"github.com/spf13/viper"
)

func Init(path string) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func GetServicePathFromIPConf() string {
	return viper.GetString("ip_conf.service_path")
}

func GetDTimeoutForDiscovery() time.Duration {
	return viper.GetDuration("discovery.timeout")
}

func GetEndpointsForDiscovery() []string {
	return viper.GetStringSlice("discovery.endpoints")
}

func GetLogLevelForLogger() string {
	return viper.GetString("global.log_level")
}

func IsDebug() bool {
	return viper.GetString("global.env") == "debug"
}

func GetCacheRedisEndpointList() []string {
	return viper.GetStringSlice("cache.redis.endpoints")
}
