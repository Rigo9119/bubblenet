package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

// Hub maneja todas las conexiones WebSocket
type Hub struct {
	// Configuración
	debug bool

	// Registro de clientes
	clients map[*Client]bool

	// Canales para comunicación
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	// WebSocket upgrader
	upgrader websocket.Upgrader
}

// NewHub crea un nuevo hub
func NewHub(debug bool) *Hub {
	return &Hub{
		debug:      debug,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// En desarrollo, aceptar cualquier origen
				// En producción, validar origins específicos
				return true
			},
		},
	}
}

// Run ejecuta el loop principal del hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Nuevo cliente se conecta
			h.clients[client] = true
			h.log("✅ Client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			// Cliente se desconecta
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.log("❌ Client disconnected. Total clients: %d", len(h.clients))
				
				// Enviar mensaje de sistema si el cliente tenía un username
				if client.username != "" {
					systemMsg := map[string]interface{}{
						"type":      "system",
						"username":  "System",
						"content":   client.username + " left the chat",
						"timestamp": time.Now(),
					}
					if msgBytes, err := json.Marshal(systemMsg); err == nil {
						h.broadcast <- msgBytes
					}
					
					// Enviar lista actualizada de usuarios
					h.BroadcastUserList()
				}
			}

		case message := <-h.broadcast:
			// Broadcast mensaje a todos los clientes
			h.log("📢 Broadcasting message to %d clients", len(h.clients))
			for client := range h.clients {
				select {
				case client.send <- message:
					// Mensaje enviado exitosamente
				default:
					// Cliente no puede recibir, desconectarlo
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

// HandleEcho maneja conexiones de echo (para testing)
func (h *Hub) HandleEcho(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}

	h.log("🔗 Echo connection established")

	// Simple echo - devolver todo lo que recibimos
	defer conn.Close()
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.log("❌ Echo WebSocket error: %v", err)
			}
			break
		}

		h.log("📨 Echo received: %s", string(message))

		// Echo back
		if err := conn.WriteMessage(messageType, message); err != nil {
			h.log("❌ Echo write error: %v", err)
			break
		}
	}
}

// HandleChat maneja conexiones de chat principal
func (h *Hub) HandleChat(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}

	// Crear nuevo cliente
	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	// Registrar cliente
	client.hub.register <- client

	// Iniciar goroutines para leer y escribir
	go client.writePump()
	go client.readPump()
}

// HandleRoom maneja conexiones a salas específicas
func (h *Hub) HandleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	if roomName == "" {
		http.Error(w, "Room name required", http.StatusBadRequest)
		return
	}

	h.log("🏠 Connection to room: %s", roomName)

	// Por ahora, usar el mismo handler que chat general
	// Más tarde implementaremos lógica de salas separadas
	h.HandleChat(w, r)
}

// GetOnlineUsers retorna la lista de usuarios conectados
func (h *Hub) GetOnlineUsers() []string {
	var users []string
	for client := range h.clients {
		if client.username != "" {
			users = append(users, client.username)
		}
	}
	return users
}

// BroadcastUserList envía la lista de usuarios a todos los clientes
func (h *Hub) BroadcastUserList() {
	users := h.GetOnlineUsers()
	userListMsg := map[string]interface{}{
		"type":      "user_list",
		"username":  "System",
		"content":   "",
		"timestamp": time.Now(),
		"users":     users,
	}
	if msgBytes, err := json.Marshal(userListMsg); err == nil {
		h.broadcast <- msgBytes
	}
}

// log helper para mensajes de debug
func (h *Hub) log(format string, args ...interface{}) {
	if h.debug {
		log.Printf("[HUB] "+format, args...)
	}
}
