package adminservice

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
)

type CreateUserRequest struct {
	UserId string `json:"user_id"`
}

func (r *CreateUserRequest) UnmarshalJSON(data []byte) error {
	type Aux CreateUserRequest
	aux := &struct {
		*Aux
	}{
		Aux: (*Aux)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		slog.Error("UNMARSHAL_CREATE_USER_REQUEST_ERROR", "error", err)
		return errors.New("INVALID_REQUEST_BODY")
	}

	missingRequiredFields := []string{}

	if aux.UserId == "" {
		missingRequiredFields = append(missingRequiredFields, "user_id")
	}

	if len(missingRequiredFields) > 0 {
		return errors.New("MISSING_REQUIRED_FIELDS: " + strings.Join(missingRequiredFields, ", "))
	}

	return nil
}
