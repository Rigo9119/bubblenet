package main

import (
	"bubblenet/internal/ui"
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Config struct {
	Room     string
	Private  bool
	Invite   bool
	Host     string
	Port     int
	Username string
}

func validateAndCreateConfig(
	room string,
	private,
	invite bool,
	host string,
	port int,
	username string,
) (ui.Config, error) {
	config := ui.Config{
		Room:     room,
		Private:  private,
		Invite:   invite,
		Host:     host,
		Port:     port,
		Username: username,
	}

	// Validaciones
	if invite && !private {
		return config, fmt.Errorf("--invite requires --private flag")
	}

	if private && room == "" {
		return config, fmt.Errorf("--private requires --room flag")
	}

	if username == "" {
		return config, fmt.Errorf("username is required, use --user flag")
	}

	if port < 1 || port > 65535 {
		return config, fmt.Errorf("invalid port: %d", port)
	}

	return config, nil
}

func main() {
	fmt.Println("Bubblenet websocket server")
	log.Println("Project initialize successfully")

	var (
		room     = flag.String("room", "", "Name of the room you want to join")
		private  = flag.Bool("private", false, "Create a private room")
		invite   = flag.Bool("invite", false, "Generate invitation code")
		host     = flag.String("host", "localhost", "Host of the server")
		port     = flag.Int("port", 8080, "Server port")
		username = flag.String("user", "", "Username")
	)

	flag.Parse()

	config, err := validateAndCreateConfig(*room, *private, *invite, *host, *port, *username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	app := ui.NewApp(config)
	program := tea.NewProgram(app, tea.WithAltScreen())

	if err := program.Start(); err != nil {
		log.Fatal("Error starting application:", err)
	}
}
