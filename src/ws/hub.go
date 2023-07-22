package ws

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"eureka/src/data"
)

const (
	SENDER_ASR = 100
	SENDER_DM  = 200
	SENDER_TTS = 300
)

type Broadcast struct {
	msgType int
	data    []byte
	sender  int
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// registered clients
	clients map[*Client]bool

	// Inbound message from the clients
	broadcast chan *Broadcast

	// Register request from the clients
	register chan *Client

	// Unregister requests from clients
	unregiser chan *Client

	// asr
	asr *AsrClient

	// tts

}

func newHub() *Hub {
	return &Hub{
		clients:   make(map[*Client]bool),
		broadcast: make(chan *Broadcast),
		register:  make(chan *Client),
		unregiser: make(chan *Client),
	}
}

func (h *Hub) run() {
	var asrIng string
	var asrResult string
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregiser:
			// client close , hub remove
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			log.Printf("client disconnect:%d, size:%d", client.sender, len(h.clients))

		case broadcast := <-h.broadcast:
			log.Printf("broadcast size:%d, sender:%d", len(broadcast.data), broadcast.sender)
			switch broadcast.sender {
			case UID_UPLOADER:
				// 音频数据发给ASR服务
				h.asr.conn.WriteMessage(broadcast.msgType, broadcast.data)
			case SENDER_ASR:
				// 识别结果
				if broadcast.msgType == websocket.TextMessage {
					asrServerText := string(broadcast.data)
					var megRes []byte
					// k2结束
					if asrServerText == "Done!" {
						asrResult = asrIng
						asrIng = ""
						result := data.AsrData{
							Topic: data.TOPIC_ASR_RESULT,
							Text:  asrResult,
						}
						megRes, _ = json.Marshal(result)

					} else {
						asrIng = asrServerText
						asrResult = ""
						result := data.AsrData{
							Topic: data.TOPIC_ASR_SPEECHING,
							Text:  asrServerText,
						}
						megRes, _ = json.Marshal(result)
					}

					//log.Print("SENDER ASR")
					for client := range h.clients {
						select {
						// 下发给到终端,Todo 修改终端端点
						case client.send <- []byte(megRes):
						default:
							close(client.send)
							delete(h.clients, client)
						}
					}

					if asrResult != "" {
						var dmReq = "{\"message\":\"" + asrResult + "\"}"
						// 请求NLP DM
						resp, err := http.Post("http://10.4.0.1:5005/webhooks/rest/webhook", "application/json", strings.NewReader(dmReq))
						if err != nil {
							log.Fatal("post dm fail, err:", err)
						}
						defer resp.Body.Close()
						resMsg, _ := ioutil.ReadAll(resp.Body)
						log.Printf("dm req:%s, res:%s", dmReq, resMsg)

						var rasaRes []data.RasaResultItem

						json.Unmarshal(resMsg, &rasaRes)

						var ttsReq = "{\"text\":\"" + rasaRes[0].Text + "\",\"spk_id\":0,\"speed\":1.0,\"volume\":1.0,\"sample_rate\":16000}"
						ttsResp, ttsErr := http.Post("http://10.4.0.1:8090/paddlespeech/tts", "application/json", strings.NewReader(ttsReq))
						if ttsErr != nil {
							log.Fatal("tts post fail, err:", err)
						}
						defer ttsResp.Body.Close()
						ttsResStr, _ := ioutil.ReadAll(ttsResp.Body)
						log.Printf("tts req:%s, res:%s", ttsReq, ttsResStr)
						var ttsData data.PaddlespeechData
						json.Unmarshal(ttsResStr, &ttsData)

						// 下发消息给到终端
						dm := data.DmData{
							Topic: data.TOPIC_DM_RESULT,
							DM: data.DmItem{
								Nlg:         rasaRes[0].Text,
								AudioBase64: ttsData.Result.Audio,
							},
						}

						dmRes, _ := json.Marshal(dm)
						//log.Print("SENDER DM")
						for client := range h.clients {
							select {
							// 下发给到终端,Todo 修改终端端点
							case client.send <- []byte(dmRes):
							default:
								close(client.send)
								delete(h.clients, client)
							}
						}
					}
				}
			}
		}
	}
}
