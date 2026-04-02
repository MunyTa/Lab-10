package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 512
)

// HandleWebSocket обрабатывает WebSocket подключение
func HandleWebSocket(hub *Hub, conn *websocket.Conn, room, user string) {
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		room: room,
		user: user,
	}

	client.hub.register <- client

	// НЕ отправляем приветственное сообщение, чтобы не мешать тестам
	// welcomeMsg удален

	// Запускаем горутины
	go client.writePump()
	go client.readPump()
}

// readPump читает сообщения от клиента
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		conn := c.conn.(*websocket.Conn)
		conn.Close()
	}()

	conn := c.conn.(*websocket.Conn)
	conn.SetReadLimit(maxMsgSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Printf("Invalid message: %v", err)
			continue
		}

		msg.User = c.user
		msg.Room = c.room
		msg.Time = time.Now().Format("15:04:05")
		c.hub.broadcast <- &msg
	}
}

// writePump отправляет сообщения клиенту
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn := c.conn.(*websocket.Conn)
		conn.Close()
	}()

	conn := c.conn.(*websocket.Conn)

	for {
		select {
		case message, ok := <-c.send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
