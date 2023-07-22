package main

import (
	"eureka/src/ws"
	"log"
	"net/http"
)

func init() {
	log.Print("init")
}

func main() {
	log.Print("server start")
	defer log.Fatal("server stop")

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
