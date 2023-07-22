package ws

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type AsrClient struct {
	hub  *Hub
	conn *websocket.Conn
}

func newAsr(h *Hub) *AsrClient {
	// asr
	u := url.URL{Scheme: "ws", Host: "10.4.0.1:6006", Path: ""}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	log.Printf("asr connecting to %s", u.String())
	if err != nil {
		log.Fatal("dial:", err)
	}
	go func() {
		defer c.Close()
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			// asr message
			log.Printf("asr recv: %s", message)
			// send asr to  hub
			h.broadcast <- &Broadcast{msgType: messageType, data: message, sender: SENDER_ASR}

		}
	}()

	return &AsrClient{
		hub:  h,
		conn: c,
	}
}
