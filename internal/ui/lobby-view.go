package ui

import "fmt"

func (m Model) LobbyView() string {
	title := titleStyle.Render("Bubblenet Lobby")

	roomsList := m.roomList.View()

	help := helpStyle.Render(
		"[↑↓] Navigate • [Enter] Join • [C] Create • [R] Refresh • [Q] Quit")

	status := statusStyle.Render(fmt.Sprintf("User: %s", m.config.Username))

	return fmt.Sprintf("%s\n%s\n\n%s\n\n%s",
		title,
		status,
		roomsList,
		help)
}
