package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/domain"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository/memory"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/service"
)

func TestCreateLobbyRejectsBlankName(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "spaces", input: "   "},
		{name: "tabs and newlines", input: "\t\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := memory.NewLobbyRepository()
			svc := service.NewLobbyService(repo)

			lobby, err := svc.CreateLobby(ctx, tt.input, 4)

			if !errors.Is(err, service.ErrInvalidLobbyName) {
				t.Errorf("error = %v, want %v", err, service.ErrInvalidLobbyName)
			}

			if lobby != nil {
				t.Errorf("lobby = %+v, want nil", lobby)
			}

			lobbies, err := repo.List(ctx)
			if err != nil {
				t.Fatalf("failed to list lobbies: %v", err)
			}

			if len(lobbies) != 0 {
				t.Errorf("len(lobbies) = %d, want 0", len(lobbies))
			}
		})
	}
}

func TestCreateLobbyRejectsInvalidMaxPlayers(t *testing.T) {
	tests := []struct {
		name       string
		maxPlayers int
	}{
		{name: "negative", maxPlayers: -1},
		{name: "zero", maxPlayers: 0},
		{name: "one", maxPlayers: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := memory.NewLobbyRepository()
			svc := service.NewLobbyService(repo)

			lobby, err := svc.CreateLobby(ctx, "Sala válida", tt.maxPlayers)

			if !errors.Is(err, service.ErrInvalidMaxPlayers) {
				t.Errorf("error = %v, want %v", err, service.ErrInvalidMaxPlayers)
			}

			if lobby != nil {
				t.Errorf("lobby = %+v, want nil", lobby)
			}

			lobbies, err := repo.List(ctx)
			if err != nil {
				t.Fatalf("failed to list lobbies: %v", err)
			}

			if len(lobbies) != 0 {
				t.Errorf("len(lobbies) = %d, want 0", len(lobbies))
			}
		})
	}
}

func TestCreateLobbySuccess(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewLobbyRepository()
	svc := service.NewLobbyService(repo)

	lobby, err := svc.CreateLobby(ctx, "  Sala de estudo  ", 2)
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}

	if lobby == nil {
		t.Fatal("lobby = nil, want created lobby")
	}

	if lobby.ID <= 0 {
		t.Errorf("ID = %d, want positive ID", lobby.ID)
	}

	if lobby.Name != "Sala de estudo" {
		t.Errorf("Name = %q, want %q", lobby.Name, "Sala de estudo")
	}

	if lobby.MaxPlayers != 2 {
		t.Errorf("MaxPlayers = %d, want 2", lobby.MaxPlayers)
	}

	if lobby.Status != domain.LobbyStatusWaiting {
		t.Errorf("Status = %q, want %q", lobby.Status, domain.LobbyStatusWaiting)
	}

	if lobby.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero, want creation time")
	}

	stored, err := repo.FindByID(ctx, lobby.ID)
	if err != nil {
		t.Fatalf("failed to find stored lobby: %v", err)
	}

	if stored == nil {
		t.Fatal("stored lobby = nil, want created lobby")
	}

	if *stored != *lobby {
		t.Errorf("stored lobby = %+v, want %+v", stored, lobby)
	}
}
