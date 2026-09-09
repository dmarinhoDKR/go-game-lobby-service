package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/domain"
)

var ErrLobbyNotFound = errors.New("lobby not found")

type LobbyRepository struct {
	mu      sync.RWMutex
	lobbies map[int64]*domain.Lobby
	nextID  int64
}

func NewLobbyRepository() *LobbyRepository {
	return &LobbyRepository{
		lobbies: make(map[int64]*domain.Lobby),
		nextID:  1,
	}
}

func (r *LobbyRepository) Create(
	ctx context.Context,
	lobby *domain.Lobby,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	lobby.ID = r.nextID
	r.nextID++

	r.lobbies[lobby.ID] = lobby

	return nil
}

func (r *LobbyRepository) FindByID(
	ctx context.Context,
	id int64,
) (*domain.Lobby, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	lobby, exists := r.lobbies[id]
	if !exists {
		return nil, ErrLobbyNotFound
	}

	return lobby, nil
}

func (r *LobbyRepository) List(
	ctx context.Context,
) ([]domain.Lobby, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	lobbies := make([]domain.Lobby, 0, len(r.lobbies))

	for _, lobby := range r.lobbies {
		lobbies = append(lobbies, *lobby)
	}

	return lobbies, nil
}
