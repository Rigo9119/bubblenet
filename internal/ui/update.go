package ui

import (
	"bubblenet/internal/client"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type reconnectMsg struct{}

// Update maneja los mensajes y actualiza el modelo
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Mensajes del sistema
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.roomList.SetSize(msg.Width-4, msg.Height-10)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	// Mensajes personalizados
	case loadCompleteMsg:
		if m.state == StateLoading && m.connectionStatus == client.StatusConnected {
			m.state = StateLobby
		}
		return m, nil

	case joinCompleteMsg:
		if m.state == StateJoining {
			m.state = StateChat
			m.currentRoom = msg.roomName
			// Agregar mensaje de sistema
			welcomeMsg := Message{
				Username:  "System",
				Content:   fmt.Sprintf("You joined #%s", msg.roomName),
				Timestamp: time.Now(),
				IsSystem:  true,
			}
			m.messages = append(m.messages, welcomeMsg)
		}
		return m, nil

	case createRoomMsg:
		m.state = StateChat
		m.currentRoom = msg.roomName
		return m, nil

	case wsConnectedMsg:
		m.connectionStatus = client.StatusConnected
		m.errorMsg = ""
		// Determinar el estado correcto según la configuración inicial
		m.state = m.config.GetInitialState()
		return m, listenForWSMessages(m.wsClient)

	case wsErrorMsg:
		m.connectionStatus = client.StatusError
		m.errorMsg = fmt.Sprintf("Connection error: %v", msg.err)
		m.state = StateError
		return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return reconnectMsg{}
		})

	case wsStatusMsg:
		fmt.Printf("DEBUG: Received wsStatusMsg with status: %v (current status: %v)\n", msg.status, m.connectionStatus)
		// Only update if it's not going backwards from Connected to Connecting
		if !(m.connectionStatus == client.StatusConnected && msg.status == client.StatusConnecting) {
			m.connectionStatus = msg.status
			fmt.Printf("DEBUG: Updated connectionStatus to: %v\n", m.connectionStatus)
		} else {
			fmt.Printf("DEBUG: Ignored status downgrade from Connected to Connecting\n")
		}
		if msg.status == client.StatusConnected {
			// Determinar el estado correcto según la configuración inicial
			m.state = m.config.GetInitialState()
			// Configurar datos iniciales para el nuevo estado
			if m.state == StateLobby {
				// Configurar lista de salas
				items := make([]list.Item, len(m.rooms))
				for i, room := range m.rooms {
					items[i] = roomItem{room}
				}
				m.roomList.SetItems(items)
			}
			return m, listenForWSMessages(m.wsClient)
		} else if msg.status == client.StatusError {
			m.state = StateError
		}
		return m, nil

	case wsMessageMsg:
		// Manejar diferentes tipos de mensajes
		switch msg.message.Type {
		case "user_list":
			// Actualizar lista de usuarios
			m.users = []User{}
			for _, username := range msg.message.Users {
				m.users = append(m.users, User{
					UserName:  username,
					UserState: "online",
				})
			}
		default:
			// Mensaje de chat normal o sistema
			newMsg := Message{
				Username:  msg.message.Username,
				Content:   msg.message.Content,
				Timestamp: msg.message.Timestamp,
				IsSystem:  msg.message.Type == "system",
			}
			m.messages = append(m.messages, newMsg)
		}

		return m, listenForWSMessages(m.wsClient)

	case reconnectMsg:
		if m.connectionStatus != client.StatusConnected {
			return m, connectWebSocket(m.wsClient)
		}
		return m, nil
	}
	// Actualizar componentes según el estado actual
	return m.updateComponents(msg)
}

// handleKeyPress maneja las teclas presionadas
func (m Model) handleKeyPress(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.state == StateLobby || m.state == StateInviting {
			return m, tea.Quit
		}
		// En chat, 'q' vuelve al lobby
		if m.state == StateChat {
			m.state = StateLobby
			m.currentRoom = ""
			m.messages = []Message{}
			return m, nil
		}

	case "esc":
		// Esc siempre vuelve al estado anterior o sale
		switch m.state {
		case StateChat, StateCreating, StateInviting:
			m.state = StateLobby
			m.currentRoom = ""
			m.messages = []Message{}
			m.errorMsg = ""
			return m, nil
		default:
			return m, tea.Quit
		}
	}

	// Manejar teclas específicas por estado
	switch m.state {
	case StateLobby:
		return m.handleLobbyKeys(msg)
	case StateChat:
		return m.handleChatKeys(msg)
	case StateInviting:
		return m.handleInvitingKeys(msg)
	case StateCreating:
		return m.handleCreatingKeys(msg)
	}

	return m, nil
}

// handleLobbyKeys maneja teclas en el lobby
func (m Model) handleLobbyKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Unirse a la sala seleccionada (solo si está conectado)
		if m.connectionStatus == client.StatusConnected && len(m.rooms) > 0 {
			selected := m.roomList.SelectedItem()
			if roomItem, ok := selected.(roomItem); ok {
				m.currentRoom = roomItem.room.Name
				m.messages = getMockMessages(roomItem.room.Name)
				m.state = StateChat
				return m, nil
			}
		}

	case "c":
		// Crear nueva sala (solo si está conectado)
		if m.connectionStatus == client.StatusConnected {
			m.state = StateCreating
			return m, nil
		}

	case "r":
		// Refrescar - reconectar si no está conectado, refrescar salas si está conectado
		if m.connectionStatus == client.StatusConnected {
			m.rooms = getMockRooms()
			items := make([]list.Item, len(m.rooms))
			for i, room := range m.rooms {
				items[i] = roomItem{room}
			}
			m.roomList.SetItems(items)
		} else {
			// Intentar reconectar
			m.connectionStatus = client.StatusConnecting
			return m, connectWebSocket(m.wsClient)
		}
		return m, nil
	}

	// Pasar navegación a la lista
	var cmd tea.Cmd
	m.roomList, cmd = m.roomList.Update(msg)
	return m, cmd
}

// handleChatKeys maneja teclas en el chat
func (m Model) handleChatKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Enviar mensaje real al servidor
		if m.messageInput.Value() != "" {
			content := m.messageInput.Value()

			// Enviar al servidor via WebSocket
			if m.connectionStatus == client.StatusConnected {
				m.wsClient.SendMessage(content)
			} else {
				// Si no hay conexión, mostrar error
				errorMsg := Message{
					Username:  "System",
					Content:   "⚠️ Not connected to server. Cannot send message.",
					Timestamp: time.Now(),
					IsSystem:  true,
				}
				m.messages = append(m.messages, errorMsg)
			}

			m.messageInput.SetValue("")
		}
		return m, nil
	}

	// Pasar input al componente de texto
	var cmd tea.Cmd
	m.messageInput, cmd = m.messageInput.Update(msg)
	return m, cmd
}

// handleInvitingKeys maneja teclas en modo invitación
func (m Model) handleInvitingKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "space":
		// Ir al chat después de mostrar la invitación
		m.state = StateChat
		return m, nil
	}
	return m, nil
}

// handleCreatingKeys maneja teclas en modo creación
func (m Model) handleCreatingKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Crear sala con el nombre ingresado
		roomName := m.messageInput.Value()
		if roomName != "" {
			return m, tea.Cmd(func() tea.Msg {
				return createRoomMsg{roomName: roomName}
			})
		}
	}

	// Pasar input al componente de texto
	var cmd tea.Cmd
	m.messageInput, cmd = m.messageInput.Update(msg)
	return m, cmd
}

// updateComponents actualiza los componentes específicos
func (m Model) updateComponents(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Actualizar componentes según el estado
	switch m.state {
	case StateLobby:
		var cmd tea.Cmd
		m.roomList, cmd = m.roomList.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case StateChat, StateCreating:
		var cmd tea.Cmd
		m.messageInput, cmd = m.messageInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}
