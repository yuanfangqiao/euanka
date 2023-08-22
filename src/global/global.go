package global

import (
	"eureka/src/config"
	"go.uber.org/zap"
	"github.com/spf13/viper"
)

var (
	CONFIG config.Configuration
	VIPER  *viper.Viper
	LOG *zap.Logger
)
