package service_test

import (
	"context"
	"errors"
	"testing"

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
