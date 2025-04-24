package exampleservice

import (
	"encoding/json"
	"testing"
)

func TestCreateExampleRequestUnmarshalJson(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		want       CreateExampleRequest
		errMessage string
	}{
		{
			name:       "valid request",
			json:       `{"message": "hello"}`,
			want:       CreateExampleRequest{Message: "hello"},
			errMessage: "",
		},
		{
			name:       "missing message",
			json:       `{"invalid": "request"}`,
			want:       CreateExampleRequest{},
			errMessage: "MISSING_REQUIRED_FIELDS: message",
		},
		{
			name:       "missing request body",
			json:       ``,
			want:       CreateExampleRequest{},
			errMessage: "unexpected end of JSON input",
		},
		{
			name:       "invalid request body",
			json:       `{"message": 123}`,
			want:       CreateExampleRequest{},
			errMessage: "INVALID_REQUEST_BODY",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var request CreateExampleRequest

			err := json.Unmarshal([]byte(tc.json), &request)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMsg)
			}
		})
	}
}

func TestPatchExampleRequestUnmarshalJson(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		want       PatchExampleRequest
		errMessage string
	}{
		{
			name:       "valid request",
			json:       `{"message": "hello"}`,
			want:       PatchExampleRequest{Message: "hello"},
			errMessage: "",
		},
		{
			name:       "missing message",
			json:       `{"invalid": "request"}`,
			want:       PatchExampleRequest{},
			errMessage: "MISSING_REQUIRED_FIELDS: message",
		},
		{
			name:       "missing request body",
			json:       ``,
			want:       PatchExampleRequest{},
			errMessage: "unexpected end of JSON input",
		},
		{
			name:       "invalid request body",
			json:       `{"message": 123}`,
			want:       PatchExampleRequest{},
			errMessage: "INVALID_REQUEST_BODY",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var request PatchExampleRequest

			err := json.Unmarshal([]byte(tc.json), &request)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMsg)
			}
		})
	}
}
