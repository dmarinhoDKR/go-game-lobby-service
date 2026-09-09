package repository

import (
	"context"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/domain"
)

type LobbyRepository interface {
	Create(ctx context.Context, lobby *domain.Lobby) error

	FindByID(ctx context.Context, id int64) (*domain.Lobby, error)

	List(ctx context.Context) ([]domain.Lobby, error)
}
