package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
