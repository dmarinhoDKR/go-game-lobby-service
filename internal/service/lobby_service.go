package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/domain"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository"
)

var (
	ErrInvalidLobbyName  = errors.New("invalid lobby name")
	ErrInvalidMaxPlayers = errors.New("invalid max players")
)

type LobbyService struct {
	repository repository.LobbyRepository
}

func NewLobbyService(repository repository.LobbyRepository) *LobbyService {
	return &LobbyService{
		repository: repository,
	}
}

func (s *LobbyService) CreateLobby(
	ctx context.Context,
	name string,
	maxPlayers int,
) (*domain.Lobby, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrInvalidLobbyName
	}

	if maxPlayers < 2 {
		return nil, ErrInvalidMaxPlayers
	}

	lobby := &domain.Lobby{
		Name:       name,
		Status:     domain.LobbyStatusWaiting,
		MaxPlayers: maxPlayers,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repository.Create(ctx, lobby); err != nil {
		return nil, fmt.Errorf("failed to create lobby: %w", err)
	}

	return lobby, nil
}

func (s *LobbyService) FindLobbyByID(
	ctx context.Context,
	id int64,
) (*domain.Lobby, error) {
	lobby, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find lobby: %w", err)
	}

	return lobby, nil
}
