package adminservice

import (
	"encoding/json"
	"testing"
)

func TestCreateUserRequestUnmarshalJson(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		want       CreateUserRequest
		errMessage string
	}{
		{
			name:       "PassingCase-ValidRequest",
			json:       `{"user_id": "a9028d39-bce2-4175-8788-e89ef91274e5"}`,
			want:       CreateUserRequest{UserId: "a9028d39-bce2-4175-8788-e89ef91274e5"},
			errMessage: "",
		},
		{
			name:       "FailingCase-MissingUserId",
			json:       `{"invalid": "request"}`,
			want:       CreateUserRequest{},
			errMessage: "MISSING_REQUIRED_FIELDS: user_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var request CreateUserRequest

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
