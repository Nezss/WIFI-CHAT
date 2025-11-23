package handlers

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"wifi-chat/models"
	"wifi-chat/storage"
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	Username string
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan models.Message
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan models.Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			// Kirim history chat ke client baru
			h.sendHistory(client)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}

		case msg := <-h.Broadcast:
			// Simpan pesan ke file
			storage.SaveMessage(msg)

			data, _ := json.Marshal(msg)
			for client := range h.Clients {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

func (h *Hub) sendHistory(client *Client) {
	messages, err := storage.LoadMessages()
	if err != nil {
		return
	}

	// Kirim sebagai history type
	historyMsg := struct {
		Type     string           `json:"type"`
		Messages []models.Message `json:"messages"`
	}{
		Type:     "history",
		Messages: messages,
	}

	data, _ := json.Marshal(historyMsg)
	client.Send <- data
}
