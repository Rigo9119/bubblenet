// internal/ui/view.go
package ui

import (
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
	return fmt.Sprintf("\n\n   %s\n\n   Connecting to server...\n\n",
		titleStyle.Render("BUBBLENET"))
}

// lobbyView muestra el lobby principal
func (m Model) lobbyView() string {
	title := titleStyle.Render("BUBBLENET LOBBY")

	// Lista de salas
	roomsList := m.roomList.View()

	// Información de ayuda
	help := helpStyle.Render(
		"[↑↓] Navigate • [Enter] Join • [C] Create • [R] Refresh • [Q] Quit")

	// Status
	status := statusStyle.Render(fmt.Sprintf("User: %s", m.config.Username))

	return fmt.Sprintf("%s\n%s\n\n%s\n\n%s",
		title,
		status,
		roomsList,
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
	// Header
	title := titleStyle.Render(fmt.Sprintf("ROOM: #%s", m.currentRoom))
	status := statusStyle.Render(fmt.Sprintf("User: %s", m.config.Username))
	header := lipgloss.JoinHorizontal(lipgloss.Top, title, " ", status)

	// Mensajes
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
	maxLines := m.height - 6 // Reservar espacio para header, input y help
	if len(messageLines) > maxLines {
		messageLines = messageLines[len(messageLines)-maxLines:]
	}

	messagesArea := strings.Join(messageLines, "\n")

	// Input de mensaje
	inputArea := fmt.Sprintf("> %s", m.messageInput.View())

	// Ayuda
	help := helpStyle.Render("[Enter] Send • [Q] Back to lobby • [Esc] Exit")

	return fmt.Sprintf("%s\n\n%s\n\n%s\n%s",
		header,
		messagesArea,
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
