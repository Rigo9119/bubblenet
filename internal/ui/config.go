package ui

type Config struct {
	Room     string
	Private  bool
	Invite   bool
	Host     string
	Port     int
	Username string
}

func (c Config) GetInitialState() AppState {
	if c.Room != "" {
		if c.Private && c.Invite {
			return StateInviting
		}
		return StateJoining
	}
	return StateLobby
}
