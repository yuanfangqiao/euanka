package ws

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/yalp/jsonpath"
)

const (
	LLM_HOST_PORT = "192.168.1.16:7600"
)

var nlgSentenceArr []string
var nlgArrIndex int 

type LlmClient struct {
	hub  *Hub
	conn *websocket.Conn
}

func newLlm(h *Hub) *LlmClient {

	u := url.URL{Scheme: "ws", Host: LLM_HOST_PORT, Path: ""}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	log.Printf("llm connecting to %s", u.String())
	if err != nil {
		log.Fatal("llm dial:", err)
	}
	go func() {
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("llm err:", err)
				return
			}

			log.Printf("llm recv: %s", message)

			h.llm.received(message)
		}
	}()

	return &LlmClient{
		hub:  h,
		conn: c,
	}

}

// 发送tts内容
func (llm *LlmClient) send(sence string, text string) {
	// 清空NLG
	nlgSentenceArr = make([]string, 0)
	nlgArrIndex = 1
	// 自定义了消息格式
	dmInputJson := `
		{
			"topic": "dm.input",
			"text": "当前在角色扮演。你扮演成游戏角色派蒙，你称呼我“旅行者”。\n场景：` + sence + ` \n根据场景回答，回答内容必须只以“派蒙：”开头。\n我问：` + text + `" 
		}
	`
	log.Printf("llm client send2:%s", dmInputJson)
	llm.conn.WriteMessage(websocket.TextMessage, []byte(dmInputJson))
}

// 收到tts结果
func (llm *LlmClient) received(message []byte) {
	/*
		{"topic": "dm.ing", "dm": {"piece": "新", "output": "蜡笔小新"}}
		{"topic": "dm.result", "dm": {"piece": "!", "output": "蜡笔小新: 今天珠海的天气真是阴沉沉的,好像随时都有下雨的样子。温度也很低,一定要小心感冒哦!"}}
	*/

	var data interface{}
	json.Unmarshal(message, &data)

	topic, errMsg := jsonpath.Read(data, "$.topic")
	if errMsg != nil {
		log.Printf("ignore errMsg:%s", errMsg)
		return
	}

	if topic == "dm.ing" {
		output, errMsg := jsonpath.Read(data, "$.dm.output")
		if errMsg != nil {
			panic(errMsg)
		}
		outputStr := output.(string)
		seps := "：:，,。.？?！!～~\n"
		nlgSentenceArr := strings.FieldsFunc(outputStr, func(r rune) bool {
			return strings.ContainsRune(seps, r)
		})
		log.Println("nlgSentenceArr:", nlgSentenceArr)
		log.Println("len:", len(nlgSentenceArr))

		if len(nlgSentenceArr) >1 && len(nlgSentenceArr) > nlgArrIndex+1 {
			nlgSentence  := nlgSentenceArr[nlgArrIndex]
			nlgArrIndex++
			llm.hub.broadcast <- &Broadcast{msgType: websocket.TextMessage, data: []byte(nlgSentence), sender: SENDER_LLM}
		}

	}

	if topic == "dm.result" {
		output, errMsg := jsonpath.Read(data, "$.dm.output")
		if errMsg != nil {
			panic(errMsg)
		}
		outputStr := output.(string)
		seps := "：:，,。.？?！!～~\n"
		nlgSentenceArr := strings.FieldsFunc(outputStr, func(r rune) bool {
			return strings.ContainsRune(seps, r)
		})
		log.Println("nlgSentenceArr:", nlgSentenceArr)
		log.Println("len:", len(nlgSentenceArr))

		nlgSentence  := nlgSentenceArr[len(nlgSentenceArr)-1]
		llm.hub.broadcast <- &Broadcast{msgType: websocket.TextMessage, data: []byte(nlgSentence), sender: SENDER_LLM}
	}

}
