package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/domain"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository/postgres"
)

func TestLobbyRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	repo := postgres.NewLobbyRepository(pool)

	lobby := &domain.Lobby{
		Name:       "Integration Test Lobby",
		Status:     domain.LobbyStatusWaiting,
		MaxPlayers: 4,
		CreatedAt:  time.Now().UTC().Truncate(time.Microsecond),
	}

	if err := repo.Create(ctx, lobby); err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cleanupCancel()

		if _, err := pool.Exec(
			cleanupCtx,
			"DELETE FROM lobbies WHERE id = $1",
			lobby.ID,
		); err != nil {
			t.Errorf("failed to clean up lobby: %v", err)
		}
	}()

	if lobby.ID <= 0 {
		t.Fatalf("ID = %d, want positive ID", lobby.ID)
	}

	checkLobby := func(got domain.Lobby) {
		t.Helper()

		if got.ID != lobby.ID ||
			got.Name != lobby.Name ||
			got.Status != lobby.Status ||
			got.MaxPlayers != lobby.MaxPlayers ||
			!got.CreatedAt.Equal(lobby.CreatedAt) {
			t.Errorf("lobby = %+v, want %+v", got, lobby)
		}
	}

	found, err := repo.FindByID(ctx, lobby.ID)
	if err != nil {
		t.Fatalf("failed to find lobby: %v", err)
	}
	if found == nil {
		t.Fatal("found lobby = nil")
	}
	checkLobby(*found)

	lobbies, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("failed to list lobbies: %v", err)
	}

	matches := 0
	for _, item := range lobbies {
		if item.ID == lobby.ID {
			matches++
			checkLobby(item)
		}
	}
	if matches != 1 {
		t.Errorf("matching lobbies = %d, want 1", matches)
	}

	missing, err := repo.FindByID(ctx, 0)
	if !errors.Is(err, repository.ErrLobbyNotFound) {
		t.Errorf("error = %v, want %v", err, repository.ErrLobbyNotFound)
	}
	if missing != nil {
		t.Errorf("missing lobby = %+v, want nil", missing)
	}
}
