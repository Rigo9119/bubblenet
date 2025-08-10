package main

import (
	"bubblenet/internal/server"
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// entry point para el servidor
func main() {
	// flags que va a manejar el CLI
	var (
		port  = flag.String("port", "8080", "Port is listening on ...")
		debug = flag.Bool("debug", false, "Enable debug mode")
	)
	flag.Parse()

	r := chi.NewRouter()
	// middleware que usa chi
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// health checks ????
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Welcome to bubblenet, server is running !!"))
	})

	// Crea el hub del websocket
	hub := server.NewHub(*debug)
	go hub.Run()

	// websockets de ejemplo
	r.Route("/ws", func(r chi.Router) {
		// Echo endpoint para testing
		r.Get("/echo", hub.HandleEcho)

		// Endpoint principal de chat
		r.Get("/chat", hub.HandleChat)

		// Endpoint por sala específica (futuro)
		r.Get("/room/{roomName}", hub.HandleRoom)
	})

	// Info de startup
	log.Printf("🚀 Bubblenet server starting on port %s", *port)
	log.Printf("📡 WebSocket endpoints:")
	log.Printf("   - Echo: ws://localhost:%s/ws/echo", *port)
	log.Printf("   - Chat: ws://localhost:%s/ws/chat", *port)
	log.Printf("🔗 Health check: http://localhost:%s/health", *port)

	// Iniciar servidor
	if err := http.ListenAndServe(":"+*port, r); err != nil {
		log.Fatal("❌ Error starting server:", err)
	}
}
