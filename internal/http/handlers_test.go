package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/domain"
	apphttp "github.com/dmarinhoDKR/go-game-lobby-service/internal/http"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository/memory"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/service"
)

func TestHealth(t *testing.T) {
	repo := memory.NewLobbyRepository()
	svc := service.NewLobbyService(repo)
	handler := apphttp.NewHandler(svc)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler.Health(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	if got := response.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", got)
	}

	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Status != "ok" {
		t.Errorf("body.Status = %s, want ok", body.Status)
	}
}

func TestCreateLobbySuccess(t *testing.T) {
	repo := memory.NewLobbyRepository()
	svc := service.NewLobbyService(repo)
	handler := apphttp.NewHandler(svc)

	body := strings.NewReader(`{"name": "Test Lobby", "max_players": 4}`)
	request := httptest.NewRequest(http.MethodPost, "/lobbies", body)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.CreateLobby(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusCreated)
	}

	if got := response.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", got)
	}

	var lobby domain.Lobby
	if err := json.NewDecoder(response.Body).Decode(&lobby); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if lobby.ID <= 0 {
		t.Errorf("ID = %d, want positive ID", lobby.ID)
	}

	if lobby.Name != "Test Lobby" {
		t.Errorf("Name = %q, want %q", lobby.Name, "Test Lobby")
	}

	if lobby.MaxPlayers != 4 {
		t.Errorf("MaxPlayers = %d, want 4", lobby.MaxPlayers)
	}

	if lobby.Status != domain.LobbyStatusWaiting {
		t.Errorf("Status = %q, want %q", lobby.Status, domain.LobbyStatusWaiting)
	}

	if lobby.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero, want creation time")
	}

	stored, err := repo.FindByID(context.Background(), lobby.ID)
	if err != nil {
		t.Fatalf("failed to find stored lobby: %v", err)
	}

	if stored == nil {
		t.Fatal("stored lobby = nil, want created lobby")
	}

	if stored.ID != lobby.ID ||
		stored.Name != lobby.Name ||
		stored.MaxPlayers != lobby.MaxPlayers ||
		stored.Status != lobby.Status ||
		!stored.CreatedAt.Equal(lobby.CreatedAt) {
		t.Errorf("stored lobby = %+v, want %+v", stored, lobby)
	}
}

func TestCreateLobbyRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "empty body",
			body: "",
		},
		{
			name: "invalid JSON",
			body: `{"name": "Test Lobby", "max_players":4`,
		},
		{
			name: "wrong field type",
			body: `{"name": "Test Lobby", "max_players": "four"}`,
		},
		{
			name: "blank name",
			body: `{"name":"   ","max_players": 4}`,
		},
		{
			name: "too few players",
			body: `{"name": "Test Lobby", "max_players":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewLobbyRepository()
			svc := service.NewLobbyService(repo)
			handler := apphttp.NewHandler(svc)

			request := httptest.NewRequest(
				http.MethodPost,
				"/lobbies",
				strings.NewReader(tt.body),
			)
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.CreateLobby(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != http.StatusBadRequest {
				t.Errorf(
					"status = %d, want %d; body = %s",
					response.StatusCode,
					http.StatusBadRequest,
					recorder.Body.String(),
				)
			}

			lobbies, err := repo.List(context.Background())
			if err != nil {
				t.Fatalf("failed to list lobbies: %v", err)
			}

			if len(lobbies) != 0 {
				t.Errorf("len(lobbies) = %d, want 0", len(lobbies))
			}
		})
	}
}

type failingLobbyRepository struct {
	err error
}

func (r *failingLobbyRepository) Create(
	ctx context.Context,
	lobby *domain.Lobby,
) error {
	return r.err
}

func (r *failingLobbyRepository) FindByID(
	ctx context.Context,
	id int64,
) (*domain.Lobby, error) {
	return nil, r.err
}

func (r *failingLobbyRepository) List(
	ctx context.Context,
) ([]domain.Lobby, error) {
	return nil, r.err
}

func TestCreateLobbyRepositoryError(t *testing.T) {
	repo := &failingLobbyRepository{
		err: errors.New("storage connection failed"),
	}
	svc := service.NewLobbyService(repo)
	handler := apphttp.NewHandler(svc)

	request := httptest.NewRequest(
		http.MethodPost,
		"/lobbies",
		strings.NewReader(`{"name": "Test Lobby", "max_players": 4}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.CreateLobby(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusInternalServerError {
		t.Errorf(
			"status = %d, want %d; body = %s",
			response.StatusCode,
			http.StatusInternalServerError,
			recorder.Body.String(),
		)
	}

	if got := recorder.Body.String(); got != "internal server error\n" {
		t.Errorf("body = %q, want %q", got, "internal server error\n")
	}
}
