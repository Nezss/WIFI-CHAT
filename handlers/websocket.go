package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"wifi-chat/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	hub.Register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		if c.Username != "" {
			c.Hub.Broadcast <- models.Message{
				Type:      "system",
				Content:   c.Username + " telah keluar dari chat",
				Timestamp: time.Now().Format("15:04"),
			}
		}
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		var msg models.Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		msg.Timestamp = time.Now().Format("15:04")

		if msg.Type == "join" {
			c.Username = msg.Username
			msg.Content = msg.Username + " bergabung ke chat"
			msg.Type = "system"
		}

		c.Hub.Broadcast <- msg
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
