package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/domain"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository/memory"
)

func TestFindByIDReturnsCopy(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewLobbyRepository()

	lobby := &domain.Lobby{
		Name:       "Sala original",
		MaxPlayers: 4,
	}

	if err := repo.Create(ctx, lobby); err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	found, err := repo.FindByID(ctx, lobby.ID)
	if err != nil {
		t.Fatalf("failed to find lobby: %v", err)
	}

	found.Name = "Sala modificada"

	again, err := repo.FindByID(ctx, lobby.ID)
	if err != nil {
		t.Fatalf("failed to find lobby again: %v", err)
	}

	if again.Name != "Sala original" {
		t.Errorf("Name = %q, want %q", again.Name, "Sala original")
	}
}

func TestCreateStoresCopy(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewLobbyRepository()

	lobby := &domain.Lobby{
		Name:       "Sala original",
		MaxPlayers: 4,
	}

	if err := repo.Create(ctx, lobby); err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	lobby.Name = "Sala modificada"

	found, err := repo.FindByID(ctx, lobby.ID)
	if err != nil {
		t.Fatalf("failed to find lobby: %v", err)
	}

	if found.Name != "Sala original" {
		t.Errorf("Name = %q, want %q", found.Name, "Sala original")
	}
}

func TestFindByIDNotFound(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewLobbyRepository()

	lobby, err := repo.FindByID(ctx, 999)

	if !errors.Is(err, repository.ErrLobbyNotFound) {
		t.Errorf("error = %v, want %v", err, repository.ErrLobbyNotFound)
	}

	if lobby != nil {
		t.Errorf("lobby = %v, want nil", lobby)
	}
}

func TestListEmpty(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewLobbyRepository()

	lobbies, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("failed to list lobbies: %v", err)
	}

	if lobbies == nil {
		t.Fatal("lobbies = nil, want initialized empty slice")
	}

	if len(lobbies) != 0 {
		t.Errorf("len(lobbies) = %d, want 0", len(lobbies))
	}
}
