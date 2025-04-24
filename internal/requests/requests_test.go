package requests

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

func TestLoadRequestBody(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given

			// When

			// Then
		})
	}
}

func TestLoadPathValue(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given

			// When

			// Then
		})
	}
}

func TestGetPaginationParameters(t *testing.T) {
	tests := []struct {
		name          string
		expectedPage  string
		expectedLimit string
		errMessage    string
	}{
		{
			name:          "PassingCase",
			expectedPage:  "1",
			expectedLimit: "10",
			errMessage:    "",
		},
		{
			name:          "FailingCase-PageTooLowError",
			expectedPage:  "-1",
			expectedLimit: "10",
			errMessage:    "page must be greater than 0",
		},
		{
			name:          "FailingCase-InvalidPage",
			expectedPage:  "null",
			expectedLimit: "10",
			errMessage:    "page must be an integer",
		},
		{
			name:          "FailingCase-LimitTooLowError",
			expectedPage:  "10",
			expectedLimit: "-1",
			errMessage:    "limit must be greater than 0",
		},
		{
			name:          "FailingCase-InvalidLimit",
			expectedPage:  "10",
			expectedLimit: "nan",
			errMessage:    "limit must be an integer",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// When
			url := fmt.Sprintf("/examples/1?limit=%s&page=%s", tc.expectedLimit, tc.expectedPage)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Errorf("Error creating request, %s", err.Error())
			}

			limit, page, err := GetPaginationParameters(req)

			// Then
			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if tc.errMessage != errMessage {
				t.Errorf("Got error %s, expected %s", errMessage, tc.errMessage)
			}

			// Only validate page + limit if we're not expecting an error
			if tc.errMessage == "" {
				pageConv, _ := strconv.Atoi(tc.expectedPage)
				if page != pageConv {
					t.Errorf("Expected page %d, got %d", pageConv, page)
				}

				limitConv, _ := strconv.Atoi(tc.expectedLimit)
				if limit != limitConv {
					t.Errorf("Expected limit %d, got %d", limitConv, limit)
				}
			}

		})
	}
}
