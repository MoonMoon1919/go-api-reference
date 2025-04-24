package healthservice

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/moonmoon1919/go-api-reference/internal/build"
	"github.com/moonmoon1919/go-api-reference/internal/responses"
)

/*
Standard health check endpoint
*/
type HealthCheckResponse struct {
	Status string `json:"status"`
	Build  string `json:"build"`
}

type HealthController struct{}

func (h HealthController) Get(w http.ResponseWriter, r *http.Request) {
	resp := HealthCheckResponse{
		Status: "ok",
		Build:  build.VERSION,
	}

	respBytes, err := json.Marshal(resp)

	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			"INTERNAL_SERVER_ERROR",
			slog.String("err", err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
	}

	responses.WriteSuccessResponse(w, &respBytes, &responses.Headers{})
}
