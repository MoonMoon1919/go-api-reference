package exampleservice

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
)

type CreateExampleRequest struct {
	Message string `json:"message"`
}

func (r *CreateExampleRequest) UnmarshalJSON(data []byte) error {
	type Aux CreateExampleRequest
	aux := &struct {
		*Aux
	}{
		Aux: (*Aux)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		slog.Error("UNMARSHAL_CREATE_EXAMPLE_REQUEST_ERROR", "error", err)
		return errors.New("INVALID_REQUEST_BODY")
	}

	missingRequiredFields := []string{}

	if aux.Message == "" {
		missingRequiredFields = append(missingRequiredFields, "message")
	}

	if len(missingRequiredFields) > 0 {
		return errors.New("MISSING_REQUIRED_FIELDS: " + strings.Join(missingRequiredFields, ", "))
	}

	return nil
}

type PatchExampleRequest struct {
	Message string `json:"message"`
}

func (r *PatchExampleRequest) UnmarshalJSON(data []byte) error {
	type Aux PatchExampleRequest
	aux := &struct {
		*Aux
	}{
		Aux: (*Aux)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		slog.Warn("UNMARSHAL_PATCH_EXAMPLE_REQUEST_ERROR", "error", err)
		return errors.New("INVALID_REQUEST_BODY")
	}

	missingRequiredFields := []string{}

	if aux.Message == "" {
		missingRequiredFields = append(missingRequiredFields, "message")
	}

	if len(missingRequiredFields) > 0 {
		return errors.New("MISSING_REQUIRED_FIELDS: " + strings.Join(missingRequiredFields, ", "))
	}

	return nil
}
