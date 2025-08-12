// internal/ui/view.go
package ui

import (
	"bubblenet/internal/client"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Estilos
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Align(lipgloss.Right)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	messageStyle = lipgloss.NewStyle().
			Padding(0, 1)

	systemMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Italic(true)

	userMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00AA00")).
				Bold(true)
)

// View renderiza la interfaz de usuario
func (m Model) View() string {
	switch m.state {
	case StateLoading:
		return m.loadingView()
	case StateLobby:
		return m.lobbyView()
	case StateJoining:
		return m.joiningView()
	case StateInviting:
		return m.invitingView()
	case StateChat:
		return m.chatView()
	case StateCreating:
		return m.creatingView()
	case StateError:
		return m.errorView()
	default:
		return "Unknown state"
	}
}

// loadingView muestra la pantalla de carga
func (m Model) loadingView() string {
	var statusText string
	var statusColor lipgloss.Color
	
	switch m.connectionStatus {
	case client.StatusConnecting:
		statusText = "Connecting to server..."
		statusColor = "#FFFF00"
	case client.StatusConnected:
		statusText = "Connected! Loading..."
		statusColor = "#00FF00"
	case client.StatusError:
		statusText = fmt.Sprintf("Connection failed: %s", m.errorMsg)
		statusColor = "#FF0000"
	default:
		statusText = "Initializing..."
		statusColor = "#888888"
	}
	
	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	
	return fmt.Sprintf("\n\n   %s\n\n   %s\n\n",
		titleStyle.Render("BUBBLENET"),
		statusStyle.Render(statusText))
}

// lobbyView muestra el lobby principal
func (m Model) lobbyView() string {
	title := titleStyle.Render("BUBBLENET LOBBY")

	// Status con información de conexión
	connectionText := m.connectionStatus.String()
	var connectionColor lipgloss.Color
	switch m.connectionStatus {
	case client.StatusConnected:
		connectionColor = "#00FF00"
	case client.StatusConnecting:
		connectionColor = "#FFFF00"
	case client.StatusError:
		connectionColor = "#FF0000"
	default:
		connectionColor = "#888888"
	}
	
	connectionStyle := lipgloss.NewStyle().Foreground(connectionColor)
	status := statusStyle.Render(fmt.Sprintf("User: %s | Status: %s", 
		m.config.Username, 
		connectionStyle.Render(connectionText)))

	// Solo mostrar lista de salas si está conectado
	var content string
	var help string
	
	if m.connectionStatus == client.StatusConnected {
		// Lista de salas disponible
		roomsList := m.roomList.View()
		help = helpStyle.Render(
			"[↑↓] Navigate • [Enter] Join • [C] Create • [R] Refresh • [Q] Quit")
		content = roomsList
	} else {
		// Mensaje de espera
		var message string
		switch m.connectionStatus {
		case client.StatusConnecting:
			message = "Connecting to server, please wait..."
		case client.StatusError:
			message = errorStyle.Render("Connection failed. Retrying...")
		default:
			message = "Initializing connection..."
		}
		help = helpStyle.Render("[R] Refresh • [Q] Quit")
		content = message
	}

	return fmt.Sprintf("%s\n%s\n\n%s\n\n%s",
		title,
		status,
		content,
		help)
}

// joiningView muestra pantalla de conexión a sala
func (m Model) joiningView() string {
	title := titleStyle.Render("JOINING ROOM")
	return fmt.Sprintf("\n\n   %s\n\n   Connecting to #%s...\n\n",
		title, m.config.Room)
}

// invitingView muestra el código de invitación
func (m Model) invitingView() string {
	title := titleStyle.Render("ROOM CREATED")

	content := fmt.Sprintf(`
   %s

   Private room '#%s' created successfully!

   Share this invite code with your friends:

   %s

   Anyone with this code can join your private room.

   %s`,
		title,
		m.config.Room,
		lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1).
			Render(m.inviteCode),
		helpStyle.Render("[Enter] Continue to chat • [Esc] Back to lobby"))

	return content
}

// chatView muestra la interfaz de chat
func (m Model) chatView() string {
	// Header simplificado
	titleText := fmt.Sprintf("ROOM: #%s", m.currentRoom)
	statusText := fmt.Sprintf("User: %s", m.config.Username)

	title := titleStyle.Render(titleText)
	status := statusStyle.Render(statusText)

	// Lista de usuarios online
	userList := ""
	if len(m.users) > 0 {
		var userNames []string
		for _, user := range m.users {
			userNames = append(userNames, "🟢 "+user.UserName)
		}
		userList = " | Online: " + strings.Join(userNames, ", ")
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, title, " ", status, userList)

	// Mensajes (mismo código que antes)
	var messageLines []string
	for _, msg := range m.messages {
		var msgStyle lipgloss.Style
		var prefix string

		if msg.IsSystem {
			msgStyle = systemMessageStyle
			prefix = "* "
		} else {
			msgStyle = messageStyle
			prefix = fmt.Sprintf("<%s> ", userMessageStyle.Render(msg.Username))
		}

		timestamp := msg.Timestamp.Format("15:04")
		line := fmt.Sprintf("[%s] %s%s",
			helpStyle.Render(timestamp),
			prefix,
			msgStyle.Render(msg.Content))
		messageLines = append(messageLines, line)
	}

	// Área de mensajes (limitamos a las últimas líneas que caben)
	maxLines := m.height - 8 // Más espacio para header expandido
	if len(messageLines) > maxLines {
		messageLines = messageLines[len(messageLines)-maxLines:]
	}

	messagesArea := strings.Join(messageLines, "\n")

	// Input de mensaje simplificado
	inputArea := fmt.Sprintf("> %s", m.messageInput.View())

	// Ayuda
	help := helpStyle.Render("[Enter] Send • [Q] Back to lobby • [Esc] Exit")

	// Mostrar error si hay
	errorArea := ""
	if m.errorMsg != "" {
		errorArea = "\n" + errorStyle.Render("⚠️ "+m.errorMsg)
	}

	return fmt.Sprintf("%s\n\n%s%s\n\n%s\n%s",
		header,
		messagesArea,
		errorArea,
		inputArea,
		help)
}

// creatingView muestra la pantalla de creación de sala
func (m Model) creatingView() string {
	title := titleStyle.Render("CREATE NEW ROOM")

	content := fmt.Sprintf(`
   %s

   Enter room name:
   %s

   %s`,
		title,
		m.messageInput.View(),
		helpStyle.Render("[Enter] Create • [Esc] Cancel"))

	return content
}

// errorView muestra errores
func (m Model) errorView() string {
	title := titleStyle.Render("ERROR")
	error := errorStyle.Render(m.errorMsg)
	help := helpStyle.Render("[Esc] Back • [Q] Quit")

	return fmt.Sprintf("\n\n   %s\n\n   %s\n\n   %s\n\n", title, error, help)
}
