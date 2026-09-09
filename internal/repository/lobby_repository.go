package repository

import (
	"context"
	"errors"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/domain"
)

var ErrLobbyNotFound = errors.New("lobby not found")

type LobbyRepository interface {
	Create(ctx context.Context, lobby *domain.Lobby) error

	FindByID(ctx context.Context, id int64) (*domain.Lobby, error)

	List(ctx context.Context) ([]domain.Lobby, error)
}
