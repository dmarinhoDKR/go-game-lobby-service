package domain

import "time"

type LobbyStatus string

const (
	LobbyStatusWaiting LobbyStatus = "waiting"
	LobbyStatusStarted LobbyStatus = "started"
	LobbyStatusClosed  LobbyStatus = "closed"
)

type Lobby struct {
	ID         int64       `json:"id"`
	Name       string      `json:"name"`
	Status     LobbyStatus `json:"status"`
	MaxPlayers int         `json:"max_players"`
	CreatedAt  time.Time   `json:"created_at"`
}
