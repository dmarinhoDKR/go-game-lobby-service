package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/dmarinhoDKR/go-game-lobby-service/internal/domain"
	apphttp "github.com/dmarinhoDKR/go-game-lobby-service/internal/http"
	"github.com/dmarinhoDKR/go-game-lobby-service/internal/repository"
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

func TestFindLobbyByIDSuccess(t *testing.T) {
	repo := memory.NewLobbyRepository()
	svc := service.NewLobbyService(repo)
	handler := apphttp.NewHandler(svc)

	created, err := svc.CreateLobby(context.Background(), "Test Lobby", 4)
	if err != nil {
		t.Fatalf("failed to create lobby: %v", err)
	}
	if created == nil {
		t.Fatal("created lobby = nil")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /lobbies/{id}", handler.FindLobbyByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/lobbies/"+strconv.FormatInt(created.ID, 10),
		nil,
	)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	if got := response.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", got)
	}

	var found domain.Lobby
	if err := json.NewDecoder(response.Body).Decode(&found); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if found.ID != created.ID ||
		found.Name != created.Name ||
		found.MaxPlayers != created.MaxPlayers ||
		found.Status != created.Status ||
		!found.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf("found lobby = %+v, want %+v", found, created)
	}
}

func TestFindLobbyByIDErrors(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		storageErr error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "non-numeric ID",
			id:         "abc",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid lobby ID\n",
		},
		{
			name:       "zero ID",
			id:         "0",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid lobby ID\n",
		},
		{
			name:       "negative ID",
			id:         "-1",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid lobby ID\n",
		},
		{
			name:       "ID overflow",
			id:         "9223372036854775808",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid lobby ID\n",
		},
		{
			name:       "not found",
			id:         "999",
			wantStatus: http.StatusNotFound,
			wantBody:   "lobby not found\n",
		},
		{
			name:       "repository failure",
			id:         "1",
			storageErr: errors.New("storage connection failed"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var repo repository.LobbyRepository = memory.NewLobbyRepository()
			if tt.storageErr != nil {
				repo = &failingLobbyRepository{err: tt.storageErr}
			}

			svc := service.NewLobbyService(repo)
			handler := apphttp.NewHandler(svc)

			mux := http.NewServeMux()
			mux.HandleFunc("GET /lobbies/{id}", handler.FindLobbyByID)

			request := httptest.NewRequest(
				http.MethodGet,
				"/lobbies/"+tt.id,
				nil,
			)
			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != tt.wantStatus {
				t.Errorf(
					"status = %d, want %d",
					response.StatusCode,
					tt.wantStatus,
				)
			}

			if got := recorder.Body.String(); got != tt.wantBody {
				t.Errorf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestListLobbiesSuccess(t *testing.T) {
	tests := []struct {
		name  string
		names []string
	}{
		{name: "empty"},
		{name: "with lobbies", names: []string{"Sala A", "Sala B"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewLobbyRepository()
			svc := service.NewLobbyService(repo)
			handler := apphttp.NewHandler(svc)

			want := make(map[int64]domain.Lobby)

			for _, name := range tt.names {
				lobby, err := svc.CreateLobby(context.Background(), name, 4)
				if err != nil {
					t.Fatalf("failed to create lobby: %v", err)
				}
				if lobby == nil {
					t.Fatal("created lobby = nil")
				}
				want[lobby.ID] = *lobby
			}

			request := httptest.NewRequest(http.MethodGet, "/lobbies", nil)
			recorder := httptest.NewRecorder()

			handler.ListLobbies(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
			}

			if got := response.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}

			var lobbies []domain.Lobby
			if err := json.NewDecoder(response.Body).Decode(&lobbies); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if lobbies == nil {
				t.Fatal("lobbies = nil, want JSON array")
			}

			if len(lobbies) != len(tt.names) {
				t.Fatalf("len(lobbies) = %d, want %d", len(lobbies), len(tt.names))
			}

			for _, lobby := range lobbies {
				expected, exists := want[lobby.ID]
				if !exists {
					t.Errorf("unexpected or duplicate lobby ID: %d", lobby.ID)
					continue
				}

				if lobby.ID != expected.ID ||
					lobby.Name != expected.Name ||
					lobby.MaxPlayers != expected.MaxPlayers ||
					lobby.Status != expected.Status ||
					!lobby.CreatedAt.Equal(expected.CreatedAt) {
					t.Errorf("lobby = %+v, want %+v", lobby, expected)
				}

				delete(want, lobby.ID)
			}

			for id := range want {
				t.Errorf("missing lobby ID: %d", id)
			}
		})
	}
}

func TestListLobbiesRepositoryError(t *testing.T) {
	repo := &failingLobbyRepository{
		err: errors.New("storage connection failed"),
	}
	svc := service.NewLobbyService(repo)
	handler := apphttp.NewHandler(svc)

	request := httptest.NewRequest(http.MethodGet, "/lobbies", nil)
	recorder := httptest.NewRecorder()

	handler.ListLobbies(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusInternalServerError {
		t.Errorf(
			"status = %d, want %d",
			response.StatusCode,
			http.StatusInternalServerError,
		)
	}

	if got := recorder.Body.String(); got != "internal server error\n" {
		t.Errorf("body = %q, want %q", got, "internal server error\n")
	}
}
