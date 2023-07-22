package ws

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type TtsClient struct {
	hub  *Hub
	conn *websocket.Conn
}

func newTts(h *Hub) *TtsClient {

	hostPort := "192.168.1.16:7886"

	// tts
	u := url.URL{Scheme: "ws", Host: hostPort, Path: "/queue/join"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	log.Printf("tts connecting to %s", u.String())
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
			// tts message
			log.Printf("tts recv: %s", message)

			/*
				{"msg":"estimation","rank":0,"queue_size":1,"avg_event_process_time":null,"avg_event_concurrent_process_time":null,"rank_eta":null,"queue_eta":1}
				{"msg":"send_data"}
				{"msg":"process_starts"}
				{"msg":"process_completed","output":{"data":["Success",{"name":"/tmp/tmpqjv2o1ix/tmpa3mq95p2.wav","data":null,"is_file":true}],"is_generating":false,"duration":5.249077558517456,"average_duration":5.249077558517456},"success":true}
			*/

			// var data interface{}
			// json.Unmarshal(message, &data)

			// msg, errMsg := jsonpath.Read(data, "$.msg")
			// if errMsg != nil {
			// 	panic(err)
			// }
			// log.Printf("")

			// send asr to  hub
			h.broadcast <- &Broadcast{msgType: messageType, data: message, sender: SENDER_TTS}

		}
	}()

	return &TtsClient{
		hub:  h,
		conn: c,
	}
}
