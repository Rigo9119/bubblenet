package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const MaxUsers = 5

type AppState int

const (
	StateLoading AppState = iota
	StateLobby
	StateJoining
	StateInviting
	StateChat
	StateCreating
	StateError
)

// String implementa fmt.Stringer para AppState
func (s AppState) String() string {
	switch s {
	case StateLoading:
		return "Loading"
	case StateLobby:
		return "Lobby"
	case StateJoining:
		return "Joining"
	case StateInviting:
		return "Inviting"
	case StateChat:
		return "Chat"
	case StateCreating:
		return "Creating"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

type LobyData struct {
	Rooms []Room
}

type ChatData struct {
	Users []User
}

type User struct {
	UserName  string
	UserState string
}

type Room struct {
	Name     string
	Users    int64
	MaxUsers int8
	Private  bool
}

type Message struct {
	Username  string
	Content   string
	Timestamp time.Time
	IsSystem  bool
}

type Model struct {
	state  AppState
	config Config

	rooms        []Room
	selectedRoom int
	roomList     list.Model

	messages     []Message
	messageInput textinput.Model

	inviteCode  string
	currentRoom string
	errorMsg    string

	width  int
	height int
}

func getMockRooms() []Room {
	return []Room{
		{Name: "general", Users: 3, MaxUsers: 5, Private: false},
		{Name: "development", Users: 1, MaxUsers: 5, Private: false},
		{Name: "gaming", Users: 4, MaxUsers: 5, Private: false},
		{Name: "random", Users: 0, MaxUsers: 5, Private: false},
	}
}

func getMockMessages(roomName string) []Message {
	now := time.Now()
	return []Message{
		{
			Username:  "System",
			Content:   fmt.Sprintf("Welcome to #%s!", roomName),
			Timestamp: now.Add(-time.Minute * 5),
			IsSystem:  true,
		},
		{
			Username:  "Alice",
			Content:   "Hello everyone!",
			Timestamp: now.Add(-time.Minute * 3),
			IsSystem:  false,
		},
		{
			Username:  "Bob",
			Content:   "Hey there! How's everyone doing?",
			Timestamp: now.Add(-time.Minute * 1),
			IsSystem:  false,
		},
	}
}

func generateInviteCode(roomName string) string {
	// En implementación real, sería un UUID o token seguro
	return fmt.Sprintf("invite-%s-%d", roomName, time.Now().Unix()%10000)
}

func NewApp(config Config) *Model {
	// config text input
	ti := textinput.New()
	ti.Placeholder = "Your message ... "
	ti.Focus()

	// rooms list config
	roomList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	roomList.Title = "Available Rooms"

	model := &Model{
		state:        config.GetInitialState(),
		config:       config,
		rooms:        getMockRooms(),
		selectedRoom: 0,
		roomList:     roomList,
		messages:     []Message{},
		messageInput: ti,
		currentRoom:  config.Room,
		inviteCode:   "",
		errorMsg:     "",
	}

	model.setupInitialData()

	return model
}

func (m *Model) setupInitialData() {
	switch m.state {
	case StateLobby:
		items := make([]list.Item, len(m.rooms))
		for i, room := range m.rooms {
			items[i] = roomItem{room}
		}
		m.roomList.SetItems(items)

	case StateJoining:
		m.messages = getMockMessages(m.config.Room)

	case StateInviting:
		m.inviteCode = generateInviteCode(m.config.Room)
		m.messages = getMockMessages(m.config.Room)
	}
}

type roomItem struct {
	room Room
}

func (r roomItem) FilterValue() string { return r.room.Name }
func (r roomItem) Title() string       { return r.room.Name }
func (r roomItem) Description() string {
	if r.room.Private {
		return fmt.Sprintf("Private • %d/%d users", r.room.Users, r.room.MaxUsers)
	}
	return fmt.Sprintf("%d/%d users", r.room.Users, r.room.MaxUsers)
}

func (m Model) Init() tea.Cmd {
	switch m.state {
	case StateLoading:
		return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
			return loadCompleteMsg{}
		})
	case StateJoining:
		return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
			return joinCompleteMsg{roomName: m.config.Room}
		})
	default:
		return nil
	}
}

type (
	loadCompleteMsg struct{}
	joinCompleteMsg struct{ roomName string }
	createRoomMsg   struct{ roomName string }
)
