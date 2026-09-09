package main

import (
	"log"
	"net/http"

	apphttp "github.com/dmarinhoDKR/go-game-lobby-service/internal/http"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository/memory"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/service"
)

func main() {
	repository := memory.NewLobbyRepository()
	lobbyService := service.NewLobbyService(repository)
	handler := apphttp.NewHandler(lobbyService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /lobbies", handler.CreateLobby)
	mux.HandleFunc("GET /lobbies/{id}", handler.FindLobbyByID)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("server running on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
