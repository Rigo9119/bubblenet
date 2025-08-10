package server

// esto manejara los clientes individuales

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Tiempo máximo para escribir mensaje
	writeWait = 10 * time.Second

	// Tiempo máximo para leer mensaje
	pongWait = 60 * time.Second

	// Intervalo de ping
	pingPeriod = (pongWait * 9) / 10

	// Tamaño máximo de mensaje
	maxMessageSize = 512
)

// Client representa una conexión WebSocket individual
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// readPump lee mensajes del WebSocket
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// Configurar límites de lectura
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Loop de lectura
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			}
			break
		}

		// Log del mensaje recibido
		c.hub.log("📨 Received from client: %s", string(message))

		// Broadcast a todos los demás clientes
		c.hub.broadcast <- message
	}
}

// writePump escribe mensajes al WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub cerró el canal
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Agregar mensajes adicionales en cola
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
