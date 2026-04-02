package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Message структура сообщения
type Message struct {
	Type    string `json:"type"`    // "chat", "join", "leave"
	Room    string `json:"room"`    // название комнаты
	User    string `json:"user"`    // имя пользователя
	Content string `json:"content"` // содержимое сообщения
	Time    string `json:"time"`    // время отправки
}

// Client WebSocket клиент
type Client struct {
	hub  *Hub
	conn interface{} // будет заменен на *websocket.Conn
	send chan []byte
	room string
	user string
}

// Hub управляет комнатами и клиентами
type Hub struct {
	rooms      map[string]map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub создает новый Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run запускает главный цикл Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[client.room]; !ok {
				h.rooms[client.room] = make(map[*Client]bool)
			}
			h.rooms[client.room][client] = true
			h.mu.Unlock()
			log.Printf("Client %s joined room %s", client.user, client.room)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.room]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					log.Printf("Client %s left room %s", client.user, client.room)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.rooms[message.Room]
			h.mu.RUnlock()

			msgBytes, _ := json.Marshal(message)
			for client := range clients {
				select {
				case client.send <- msgBytes:
				default:
					close(client.send)
					delete(clients, client)
				}
			}
		}
	}
}
