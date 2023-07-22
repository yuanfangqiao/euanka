package ws

import (
	"bytes"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func (up *Client) readUploaderPump() {
	defer func() {
		up.hub.unregiser <- up
		up.conn.Close()
	}()

	up.conn.SetReadLimit(maxMessageSize)
	up.conn.SetReadDeadline(time.Now().Add(pongWait))
	up.conn.SetPongHandler(func(appData string) error { up.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		msgType, message, err := up.conn.ReadMessage()
		if err != nil {
			log.Print("upload err")
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		switch msgType {
		case websocket.TextMessage:
			message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
			//var text = string(message)
			// up.hub.broadcast <- &Broadcast{data: message}
			//log.Print("rev text:" + text)
			up.hub.broadcast <- &Broadcast{msgType: websocket.TextMessage, data: message, sender: up.sender}
			break
		case websocket.BinaryMessage:
			// binary message
			// log.Printf("rev binary size:%d", len(message))
			// send to hub
			up.hub.broadcast <- &Broadcast{msgType: websocket.BinaryMessage, data: message, sender: up.sender}
			break
		case websocket.CloseMessage:
			log.Print("rev close")
			break
		}
	}
}

func (c *Client) writeUploaderPump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			log.Print("writeUploaderPump ing")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Fatal("error:", err)
				return
			}

			w.Write(message)
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				log.Fatal("websocekt writer err:", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Fatal("send ping err:", err)
				return
			}
		}
	}
}
