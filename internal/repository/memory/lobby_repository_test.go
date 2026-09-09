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

func TestListReturnsAllLobbies(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewLobbyRepository()

	first := &domain.Lobby{Name: "Sala A", MaxPlayers: 4}
	second := &domain.Lobby{Name: "Sala B", MaxPlayers: 2}

	for _, lobby := range []*domain.Lobby{first, second} {
		if err := repo.Create(ctx, lobby); err != nil {
			t.Fatalf("failed to create lobby: %v", err)
		}
	}

	lobbies, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("failed to list lobbies: %v", err)
	}

	if len(lobbies) != 2 {
		t.Fatalf("len(lobbies) = %d, want 2", len(lobbies))
	}

	want := map[int64]domain.Lobby{
		first.ID:  *first,
		second.ID: *second,
	}

	for _, lobby := range lobbies {
		expected, exists := want[lobby.ID]
		if !exists {
			t.Errorf("unexpected or duplicate lobby ID %d", lobby.ID)
			continue
		}

		if lobby != expected {
			t.Errorf("lobby = %+v, want %+v", lobby, expected)
		}

		delete(want, lobby.ID)
	}

	for id := range want {
		t.Errorf("missing lobby ID %d", id)
	}
}

func TestListReturnsCopies(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewLobbyRepository()

	lobby := &domain.Lobby{
		Name:       "Sala original",
		MaxPlayers: 4,
	}

	if err := repo.Create(ctx, lobby); err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	lobbies, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("failed to list lobbies: %v", err)
	}

	if len(lobbies) != 1 {
		t.Fatalf("len(lobbies) = %d, want 1", len(lobbies))
	}

	lobbies[0].Name = "Sala modificada"

	found, err := repo.FindByID(ctx, lobby.ID)
	if err != nil {
		t.Fatalf("failed to find lobby: %v", err)
	}

	if found.Name != "Sala original" {
		t.Errorf("Name = %q, want %q", found.Name, "Sala original")
	}
}
