package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/service"
)

type Handler struct {
	lobbyService *service.LobbyService
}

func NewHandler(lobbyService *service.LobbyService) *Handler {
	return &Handler{
		lobbyService: lobbyService,
	}
}

type createLobbyRequest struct {
	Name       string `json:"name"`
	MaxPlayers int    `json:"max_players"`
}

func (h *Handler) CreateLobby(w http.ResponseWriter, r *http.Request) {
	var request createLobbyRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	lobby, err := h.lobbyService.CreateLobby(
		r.Context(),
		request.Name,
		request.MaxPlayers,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidLobbyName):
			http.Error(w, "invalid lobby name", http.StatusBadRequest)

		case errors.Is(err, service.ErrInvalidMaxPlayers):
			http.Error(w, "invalid max players", http.StatusBadRequest)

		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(lobby); err != nil {
		return
	}
}
