package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	UID_UPLOADER = 1
)

var UID_CLIENT = UID_UPLOADER

func uploaderWs(w http.ResponseWriter, r *http.Request, hub *Hub) {
	log.Print("upload ws setting1")
	// upgrader
	conn, err := bwUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// build client
	uploader := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), sender: UID_UPLOADER}
	uploader.hub.register <- uploader
	log.Print("upload ws setting")
	go uploader.writeUploaderPump()
	go uploader.readUploaderPump()
}

var bwUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SetUpHub() *Hub {

	hub := newHub()

	asr := newAsr(hub)
	hub.asr = asr

	llm := newLlm(hub)
	hub.llm = llm

	tts := newTts(hub)
	hub.tts = tts

	go hub.run()
	return hub
}

func SetUploader(serverMux *http.ServeMux, hub *Hub) {
	log.Printf("set upload")
	serverMux.HandleFunc("/ws/chat", func(w http.ResponseWriter, r *http.Request) {
		uploaderWs(w, r, hub)
	})
}
