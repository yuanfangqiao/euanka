package main

import (
	"eureka/src/core"
	"eureka/src/global"
	"eureka/src/ws"
	"net/http"

	"go.uber.org/zap"
)

func init() {
	// log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
	// log.Print("init")
	//global.LOG.Info("init")
}

func main() {

	//	TODO：1.配置初始化
	global.VIPER = core.InitializeViper()

	//	TODO：2.日志
	global.LOG = core.InitializeZap()
	zap.ReplaceGlobals(global.LOG)
	global.LOG.Info("server start")
	global.LOG.Info("server run success on ", zap.String("zap_log", "zap_log"), zap.Any("",global.CONFIG.Service.Asr))
	defer global.LOG.Fatal("server stop")

	serverMux := http.NewServeMux()

	var hub *ws.Hub
	hub = ws.SetUpHub()

	ws.SetUploader(serverMux, hub)
	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}
	server.ListenAndServe()
}
