package healthservice

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthController_Get(t *testing.T) {
	// GIVEN
	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	constroller := HealthController{}

	// WHEN
	constroller.Get(wr, req)

	// THEN
	if wr.Code != http.StatusOK {
		t.Fatalf("expected status code to be %d, got %d", http.StatusOK, wr.Code)
	}

	var response HealthCheckResponse
	err := json.Unmarshal(wr.Body.Bytes(), &response)

	if err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if response.Status != "ok" {
		t.Fatalf("expected status to be ok, got %s", response.Status)
	}

	if response.Build == "" {
		t.Fatalf("expected build to be set, got empty string")
	}
}
