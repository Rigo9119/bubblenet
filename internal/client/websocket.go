package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient maneja la conexión WebSocket
type WSClient struct {
	conn     *websocket.Conn
	url      string
	username string
	debug    bool

	// Canales para comunicación con la UI
	incoming chan WSMessage
	outgoing chan WSMessage
	errors   chan error
	status   chan ConnectionStatus
}

// ConnectionStatus representa el estado de la conexión
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusError
)

func (s ConnectionStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "Disconnected"
	case StatusConnecting:
		return "Connecting"
	case StatusConnected:
		return "Connected"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// WSMessage representa un mensaje WebSocket
type WSMessage struct {
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Room      string    `json:"room,omitempty"`
	Users     []string  `json:"users,omitempty"` // Para mensajes de tipo user_list
}

// NewWSClient crea un nuevo cliente WebSocket
func NewWSClient(host string, port int, username string, debug bool) *WSClient {
	wsURL := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/ws/chat",
	}

	return &WSClient{
		url:      wsURL.String(),
		username: username,
		debug:    debug,
		incoming: make(chan WSMessage, 100),
		outgoing: make(chan WSMessage, 100),
		errors:   make(chan error, 10),
		status:   make(chan ConnectionStatus, 10),
	}
}

// Connect establece la conexión WebSocket
func (ws *WSClient) Connect() error {
	ws.log("🔗 Connecting to %s", ws.url)
	ws.status <- StatusConnecting

	conn, _, err := websocket.DefaultDialer.Dial(ws.url, nil)
	if err != nil {
		ws.log("❌ Connection failed: %v", err)
		ws.status <- StatusError
		ws.errors <- err
		return err
	}

	ws.conn = conn
	ws.log("✅ Connected successfully")
	ws.log("🔄 Sending StatusConnected to status channel")
	ws.status <- StatusConnected
	ws.log("✅ StatusConnected sent to status channel")

	// Iniciar goroutines de lectura y escritura
	go ws.readLoop()
	go ws.writeLoop()

	return nil
}

// SendMessage envía un mensaje
func (ws *WSClient) SendMessage(content string) {
	message := WSMessage{
		Type:      "chat",
		Username:  ws.username,
		Content:   content,
		Timestamp: time.Now(),
	}

	select {
	case ws.outgoing <- message:
		ws.log("📤 Queued message: %s", content)
	default:
		ws.log("⚠️ Outgoing queue full, dropping message")
	}
}

// GetIncomingChannel retorna el canal de mensajes entrantes
func (ws *WSClient) GetIncomingChannel() <-chan WSMessage {
	return ws.incoming
}

// GetErrorChannel retorna el canal de errores
func (ws *WSClient) GetErrorChannel() <-chan error {
	return ws.errors
}

// GetStatusChannel retorna el canal de estado
func (ws *WSClient) GetStatusChannel() <-chan ConnectionStatus {
	return ws.status
}

// Close cierra la conexión
func (ws *WSClient) Close() {
	if ws.conn != nil {
		ws.log("🔌 Closing connection")
		ws.conn.Close()
		ws.status <- StatusDisconnected
	}
}

// readLoop lee mensajes del servidor
func (ws *WSClient) readLoop() {
	defer ws.conn.Close()

	for {
		_, messageBytes, err := ws.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ws.log("❌ Read error: %v", err)
				ws.errors <- err
			}
			ws.status <- StatusDisconnected
			return
		}

		// Intentar parsear como JSON
		var message WSMessage
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			// Si no es JSON válido, tratarlo como mensaje de texto simple
			message = WSMessage{
				Type:      "chat",
				Username:  "Unknown",
				Content:   string(messageBytes),
				Timestamp: time.Now(),
			}
		}

		ws.log("📥 Received: %s", message.Content)

		select {
		case ws.incoming <- message:
		default:
			ws.log("⚠️ Incoming queue full, dropping message")
		}
	}
}

// writeLoop escribe mensajes al servidor
func (ws *WSClient) writeLoop() {
	ticker := time.NewTicker(54 * time.Second) // Ping cada 54 segundos
	defer ticker.Stop()

	for {
		select {
		case message := <-ws.outgoing:
			ws.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			// Enviar como JSON
			if err := ws.conn.WriteJSON(message); err != nil {
				ws.log("❌ Write error: %v", err)
				ws.errors <- err
				return
			}

			ws.log("📤 Sent: %s", message.Content)

		case <-ticker.C:
			ws.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				ws.log("❌ Ping error: %v", err)
				return
			}
		}
	}
}

// log helper
func (ws *WSClient) log(format string, args ...interface{}) {
	if ws.debug {
		log.Printf("[WS] "+format, args...)
	}
}
