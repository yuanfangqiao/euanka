package core

import (
	"eureka/src/core/internal"
	"eureka/src/global"

	"flag"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// InitializeViper 优先级: 命令行 > 环境变量 > 默认值
func InitializeViper(path ...string) *viper.Viper {
	var config string

	if len(path) == 0 {
		// 定义命令行flag参数，格式：flag.TypeVar(Type指针, flag名, 默认值, 帮助信息)
		flag.StringVar(&config, "c", "", "choose config file.")

		// 定义好命令行flag参数后，需要通过调用flag.Parse()来对命令行参数进行解析。
		flag.Parse()

		// 判断命令行参数是否为空
		if config == "" {
			/*
				判断 internal.ConfigEnv 常量存储的环境变量是否为空
				比如我们启动项目的时候，执行：GVA_CONFIG=config.yaml go run main.go
				这时候 os.Getenv(internal.ConfigEnv) 得到的就是 config.yaml
				当然，也可以通过 os.Setenv(internal.ConfigEnv, "config.yaml") 在初始化之前设置
			*/
			if configEnv := os.Getenv(internal.ConfigEnv); configEnv == "" {
				switch gin.Mode() {
				case gin.DebugMode:
					config = internal.ConfigDefaultFile
					fmt.Printf("您正在使用gin模式的%s环境名称,config的路径为%s\n", gin.EnvGinMode, internal.ConfigDefaultFile)
				case gin.ReleaseMode:
					config = internal.ConfigReleaseFile
					fmt.Printf("您正在使用gin模式的%s环境名称,config的路径为%s\n", gin.EnvGinMode, internal.ConfigReleaseFile)
				case gin.TestMode:
					config = internal.ConfigTestFile
					fmt.Printf("您正在使用gin模式的%s环境名称,config的路径为%s\n", gin.EnvGinMode, internal.ConfigTestFile)
				}
			} else {
				// internal.ConfigEnv 常量存储的环境变量不为空 将值赋值于config
				config = configEnv
				fmt.Printf("您正在使用%s环境变量,config的路径为%s\n", internal.ConfigEnv, config)
			}
		} else {
			// 命令行参数不为空 将值赋值于config
			fmt.Printf("您正在使用命令行的-c参数传递的值,config的路径为%s\n", config)
		}
	} else {
		// 函数传递的可变参数的第一个值赋值于config
		config = path[0]
		fmt.Printf("您正在使用func Viper()传递的值,config的路径为%s\n", config)
	}

	fmt.Println("mode: ", gin.Mode())

	vip := viper.New()
	vip.SetConfigFile(config)
	vip.SetConfigType("yaml")

	err := vip.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// 替换 ${HOST} ，取环境变量
	for _, k := range vip.AllKeys() {
		v := vip.GetString(k)
		vip.Set(k, os.ExpandEnv(v))
	}

	vip.WatchConfig()

	vip.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err = vip.Unmarshal(&global.CONFIG); err != nil {
			fmt.Println(err)
		}
	})

	if err = vip.Unmarshal(&global.CONFIG); err != nil {
		fmt.Println(err)
	}

	fmt.Println("====1-viper====: viper init config success")

	return vip
}
