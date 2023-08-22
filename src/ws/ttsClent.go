package ws

import (
	"encoding/json"
	"eureka/src/global"
	"log"
	"net/url"
	"strconv"
	"github.com/gorilla/websocket"
	"github.com/yalp/jsonpath"
	"go.uber.org/zap"
	"sync"
	"time"
)

// const (
// 	// 一定要终端能访问的
// 	TTS_HOST_PORT = "192.168.1.16:7860"
// )

var APP_EXTERNAL_TTS_HOST_PORT string
// not asr ing asrState=0, asr ing asrState=1 
var ttsState int = 0
var curNlg string
var ttsChan chan string = make(chan string,100)
var mu sync.Mutex

type TtsClient struct {
	hub  *Hub
	conn *websocket.Conn
}

var session_hash int = 100

func newTts(h *Hub) *TtsClient {

	// 
	APP_EXTERNAL_TTS_HOST_PORT = global.CONFIG.App.ExternalTtsHostPort
	global.LOG.Info("app external host port", zap.String("APP_EXTERNAL_HOST_PORT", APP_EXTERNAL_TTS_HOST_PORT))

	// tts
	ttsHost := global.CONFIG.Service.Llm
	global.LOG.Info("tts service", zap.String("ttsHost", ttsHost))

	u := url.URL{Scheme: "ws", Host: ttsHost + ":7860", Path: "/queue/join"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	log.Printf("tts connecting to %s", u.String())
	if err != nil {
		log.Fatal("tts dial:", err)
	}
	go func() {
		defer c.Close()
		for {
			 _, message, err := c.ReadMessage()
			if err != nil {
				log.Println("tts err:", err)
				return
			}
			// tts message
			log.Printf("tts recv: %s", message)

			// var data interface{}
			// json.Unmarshal(message, &data)

			// msg, errMsg := jsonpath.Read(data, "$.msg")
			// if errMsg != nil {
			// 	panic(err)
			// }
			//log.Printf("")

			// send asr to  hub
			//h.broadcast <- &Broadcast{msgType: messageType, data: message, sender: SENDER_TTS}
			h.tts.received(message)
		}
	}()

	go func() {
		for {
			// select{
			// case nlg := <- ttsChan :
			// 	//mu.Lock()
			// 	ttsState = 1
			// 	global.LOG.Info("send nlg to tts",zap.String("nlg",nlg))
			// 	h.tts.sendTts(nlg)
			// 	for ttsState == 1{
			// 		time.Sleep(time.Millisecond * 100)
			// 	}
			// }
			nlg := <- ttsChan
			mu.Lock()
			h.tts.sendTts(nlg)
			global.LOG.Info("send nlg to tts",zap.String("nlg",nlg))
		
		}
	}()

	return &TtsClient{
		hub:  h,
		conn: c,
	}
}

// 发送tts内容
func (tts *TtsClient) send(nlg string) {

	// 放入队列
	global.LOG.Info("add nlg to chan",zap.String("nlg",nlg))
	ttsChan <- nlg
	
}

// 发送到asr服务
func (tts *TtsClient) sendTts(nlg string) {
	
	curNlg = nlg
	// 先发
	session_hash++
	session_hash_str := "1qo1fkbewlx" + strconv.Itoa(session_hash)
	fn := `{"fn_index":2,"session_hash":"` + session_hash_str + `"}`
	writeErr := tts.conn.WriteMessage(websocket.TextMessage, []byte(fn))

	if writeErr == websocket.ErrCloseSent || websocket.IsCloseError(writeErr, websocket.CloseNormalClosure) {
		log.Println("tts close reconnect")
		tts = newTts(tts.hub)
		tts.hub.tts = tts
		fn := `{"fn_index":2,"session_hash":"` + session_hash_str + `"}`
		writeErr = tts.conn.WriteMessage(websocket.TextMessage, []byte(fn))
		if writeErr != nil {
			log.Println("tts close reconnect fail", writeErr)
			return
		}
	}

	log.Printf("tts client send:%s", fn)
	fnNlg := `{"data":["` + nlg + `","派蒙 Paimon (Genshin Impact)","简体中文",1,false],"event_data":null,"fn_index":2,"session_hash":"` + session_hash_str + `"}`
	log.Printf("tts client send2:%s", fnNlg)
	tts.conn.WriteMessage(websocket.TextMessage, []byte(fnNlg))
}

// 收到tts结果
func (tts *TtsClient) received(message []byte) {

	
	/*
		{"msg":"estimation","rank":0,"queue_size":1,"avg_event_process_time":null,"avg_event_concurrent_process_time":null,"rank_eta":null,"queue_eta":1}
		{"msg":"send_data"}
		{"msg":"process_starts"}
		{"msg":"process_completed","output":{"data":["Success",{"name":"/tmp/tmpqjv2o1ix/tmpa3mq95p2.wav","data":null,"is_file":true}],"is_generating":false,"duration":5.249077558517456,"average_duration":5.249077558517456},"success":true}
	*/

	var data interface{}
	json.Unmarshal(message, &data)

	jmsg, errMsg := jsonpath.Read(data, "$.msg")
	if errMsg != nil {
		//panic(errMsg)
		log.Printf("tts received err:", errMsg)
		return
	}

	if jmsg == "process_completed" {

		fileName, errMsg := jsonpath.Read(data, "$.output.data[1].name")
		if errMsg != nil {
			panic(errMsg)
		}
		fileNameStr := fileName.(string)

		// 音频下载地址
		downloadUrl := "http://" + APP_EXTERNAL_TTS_HOST_PORT + "/file=" + fileNameStr

		ttsRes := `{"nlg":"` + curNlg +`","audioUrl":"`+ downloadUrl+`"}`

		tts.hub.broadcast <- &Broadcast{msgType: websocket.TextMessage, data: []byte(ttsRes), sender: SENDER_TTS}

		// 当前识别结束
		global.LOG.Info("tts sentence res",zap.String("curNlg",curNlg), zap.String("audioUrl", downloadUrl))
		ttsState = 0
		time.Sleep(200 * time.Millisecond)
		mu.Unlock()
	}
}
