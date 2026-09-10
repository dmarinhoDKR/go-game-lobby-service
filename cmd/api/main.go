package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	apphttp "github.com/dmarinhoDKR/go-game-lobby-service/internal/http"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository/postgres"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	cancel()

	repository := postgres.NewLobbyRepository(pool)
	lobbyService := service.NewLobbyService(repository)
	handler := apphttp.NewHandler(lobbyService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /lobbies", handler.CreateLobby)
	mux.HandleFunc("GET /lobbies/{id}", handler.FindLobbyByID)
	mux.HandleFunc("GET /lobbies", handler.ListLobbies)
	mux.HandleFunc("GET /health", handler.Health)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("server running on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("HTTP server stopped: %w", err)
	}

	return nil
}
