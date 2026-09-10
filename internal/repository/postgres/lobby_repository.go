package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/domain"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository"
)

type LobbyRepository struct {
	pool *pgxpool.Pool
}

var _ repository.LobbyRepository = (*LobbyRepository)(nil)

func NewLobbyRepository(pool *pgxpool.Pool) *LobbyRepository {
	return &LobbyRepository{
		pool: pool,
	}
}

func (r *LobbyRepository) Create(
	ctx context.Context,
	lobby *domain.Lobby,
) error {
	const query = `
		INSERT INTO lobbies (name, status, max_players, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		lobby.Name,
		string(lobby.Status),
		lobby.MaxPlayers,
		lobby.CreatedAt,
	).Scan(&lobby.ID)
	if err != nil {
		return fmt.Errorf("failed to insert lobby: %w", err)
	}

	return nil
}

func (r *LobbyRepository) FindByID(
	ctx context.Context,
	id int64,
) (*domain.Lobby, error) {
	const query = `
		SELECT id, name, status, max_players, created_at
		FROM lobbies
		WHERE id = $1
	`

	var lobby domain.Lobby
	var status string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&lobby.ID,
		&lobby.Name,
		&status,
		&lobby.MaxPlayers,
		&lobby.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrLobbyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find lobby: %w", err)
	}

	lobby.Status = domain.LobbyStatus(status)

	return &lobby, nil
}

func (r *LobbyRepository) List(
	ctx context.Context,
) ([]domain.Lobby, error) {
	const query = `
		SELECT id, name, status, max_players, created_at
		FROM lobbies
		ORDER BY id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list lobbies: %w", err)
	}
	defer rows.Close()

	lobbies := make([]domain.Lobby, 0)

	for rows.Next() {
		var lobby domain.Lobby
		var status string

		if err := rows.Scan(
			&lobby.ID,
			&lobby.Name,
			&status,
			&lobby.MaxPlayers,
			&lobby.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan lobby: %w", err)
		}

		lobby.Status = domain.LobbyStatus(status)
		lobbies = append(lobbies, lobby)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate lobbies: %w", err)
	}

	return lobbies, nil
}
