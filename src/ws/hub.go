package ws

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"github.com/yalp/jsonpath"
	"eureka/src/data"
	"eureka/src/global"
)

const (
	SENDER_ASR = 100
	SENDER_DM  = 200
	SENDER_LLM = 201
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

	// llm
	llm *LlmClient

	// tts
	tts *TtsClient
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

	var nlg string
	var nlgSentence string

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

				// k2 使用float32
				var res []byte
				if broadcast.msgType == websocket.BinaryMessage {
					res = converPcm16ToFloat32(broadcast.data)
				} else {
					global.LOG.Info("asr upload text:", zap.String("text", string(broadcast.data)))
					res = broadcast.data
				}

				//h.wsClient.conn.WriteMessage(websocket.BinaryMessage, res)
				h.asr.conn.WriteMessage(broadcast.msgType, res)
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

						rasaHost := global.CONFIG.Service.Rasa

						global.LOG.Info("rasa service", zap.String("rasaHost", rasaHost))

						// 请求NLP DM
						resp, err := http.Post("http://"+rasaHost+":5005/webhooks/rest/webhook", "application/json", strings.NewReader(dmReq))
						if err != nil {
							log.Fatal("post dm fail, err:", err)
						}
						defer resp.Body.Close()
						resMsg, _ := ioutil.ReadAll(resp.Body)
						log.Printf("dm req:%s, res:%s", dmReq, resMsg)

						var rasaRes []data.RasaResultItem

						json.Unmarshal(resMsg, &rasaRes)

						nlg = rasaRes[0].Text

						// 送LLM 人设对话
						h.llm.send(nlg, asrResult)

						// 直接播报TTS
						// h.tts.send(nlg)

						// var ttsReq = "{\"text\":\"" + rasaRes[0].Text + "\",\"spk_id\":0,\"speed\":1.0,\"volume\":1.0,\"sample_rate\":16000}"
						// ttsResp, ttsErr := http.Post("http://10.4.0.1:8090/paddlespeech/tts", "application/json", strings.NewReader(ttsReq))
						// if ttsErr != nil {
						// 	log.Fatal("tts post fail, err:", err)
						// }
						// defer ttsResp.Body.Close()
						// ttsResStr, _ := ioutil.ReadAll(ttsResp.Body)
						// log.Printf("tts req:%s, res:%s", ttsReq, ttsResStr)
						// var ttsData data.PaddlespeechData
						// json.Unmarshal(ttsResStr, &ttsData)

						// // 下发消息给到终端
						// dm := data.DmData{
						// 	Topic: data.TOPIC_DM_RESULT,
						// 	DM: data.DmItem{
						// 		Nlg:         rasaRes[0].Text,
						// 		AudioBase64: ttsData.Result.Audio,
						// 	},
						// }

						// dmRes, _ := json.Marshal(dm)
						// //log.Print("SENDER DM")
						// for client := range h.clients {
						// 	select {
						// 	// 下发给到终端,Todo 修改终端端点
						// 	case client.send <- []byte(dmRes):
						// 	default:
						// 		close(client.send)
						// 		delete(h.clients, client)
						// 	}
						// }

					}
				}
			case SENDER_LLM:
				nlgSentence = string(broadcast.data)
				log.Println("llm to tts:", string(nlgSentence))
				h.tts.send(string(nlgSentence))
			case SENDER_TTS:
				// audioUrl := string(broadcast.data)

				// ttsRes := string(broadcast.data)

				var ttsResJ interface{}
				json.Unmarshal(broadcast.data, &ttsResJ)

				nlgT, _ := jsonpath.Read(ttsResJ, "$.nlg")
				var nlg = nlgT.(string)

				audioUrlT, _ := jsonpath.Read(ttsResJ, "$.audioUrl")
				var audioUrl = audioUrlT.(string)
				

				// 下发消息给到终端
				dm := data.DmData{
					Topic: data.TOPIC_DM_RESULT,
					DM: data.DmItem{
						Nlg:         nlg,
						AudioBase64: "",
						AudioUrl:    audioUrl,
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

// 将pcm16int 转 float32 , k2需要
func converPcm16ToFloat32(pcmI16 []byte) []byte {

	res := make([]byte, 0)
	end := 0
	for i := 0; i < len(pcmI16); {
		if i+2 < len(pcmI16) {
			end = i + 2
		} else {
			end = len(pcmI16)
		}

		//itemInt16 := binary.LittleEndian.int16(before[i:end])

		//itemInt16 := int(binary.LittleEndian.Uint16(before[i:end]))

		binBuf := bytes.NewBuffer(pcmI16[i:end])

		var x int16
		binary.Read(binBuf, binary.LittleEndian, &x)

		//fmt.Printf("-%X", x)

		itemFloat32 := float32(x) / 32768

		b := Float32ToByte(itemFloat32)

		res = append(res, b...)

		i += 2
	}
	return res
}

func Float32ToByte(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}
