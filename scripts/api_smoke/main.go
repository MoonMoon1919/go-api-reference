package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const URL_BASE = "http://localhost:8080"

type CreateExampleRequest struct {
	Message string `json:"message"`
}

type PatchExampleRequest struct {
	Message string `json:"message"`
}

type CreateExampleResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

type GetExampleResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

type PatchExampleResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

// CLIENT
func createExample(message string) CreateExampleResponse {
	body, _ := json.Marshal(CreateExampleRequest{
		Message: message,
	})

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/examples", URL_BASE), bytes.NewBuffer(body))

	if err != nil {
		slog.LogAttrs(context.TODO(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.LogAttrs(context.TODO(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		slog.LogAttrs(context.TODO(), slog.LevelError, "ERROR", slog.String("err", fmt.Sprintf("Expected status 201, got %d", resp.StatusCode)))
	}

	var respBody CreateExampleResponse
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		slog.LogAttrs(context.TODO(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	return respBody
}

func getExample(id string) GetExampleResponse {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/examples/%s", URL_BASE, id), nil)
	if err != nil {
		slog.LogAttrs(context.TODO(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.LogAttrs(context.TODO(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.LogAttrs(context.TODO(), slog.LevelError, "ERROR", slog.String("err", fmt.Sprintf("expected status 200, got %d", resp.StatusCode)))
	}

	var respBody GetExampleResponse
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		slog.LogAttrs(context.TODO(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	return respBody
}

func patchExample(id string, message string) PatchExampleResponse {
	body, _ := json.Marshal(PatchExampleRequest{
		Message: message,
	})

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/examples/%s", URL_BASE, id), bytes.NewBuffer(body))
	if err != nil {
		slog.LogAttrs(req.Context(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.LogAttrs(req.Context(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.LogAttrs(req.Context(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	var respBody PatchExampleResponse
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		slog.LogAttrs(req.Context(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	return respBody
}

func deleteExample(id string) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/examples/%s", URL_BASE, id), nil)
	if err != nil {
		slog.LogAttrs(req.Context(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.LogAttrs(req.Context(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}

	if resp.StatusCode != http.StatusNoContent {
		slog.LogAttrs(req.Context(), slog.LevelError, "ERROR", slog.String("err", err.Error()))
	}
}

func main() {
	exampleIds := []string{}

	// Create 1000 examples
	for i := range 1000 {
		resp := createExample(fmt.Sprintf("message-%d", i))

		exampleIds = append(exampleIds, resp.Id)

		time.Sleep(10 * time.Millisecond)
	}

	// Retrieve 1000 examples w/o cache
	for _, id := range exampleIds {
		getExample(id)

		time.Sleep(10 * time.Millisecond)
	}

	// Patch 1000 examples
	for i, id := range exampleIds {
		patchExample(id, fmt.Sprintf("message-nu-%d", i))

		time.Sleep(10 * time.Millisecond)
	}

	// GET 1000 examples and compare value
	for i, id := range exampleIds {
		resp := getExample(id)

		if resp.Message != fmt.Sprintf("message-nu-%d", i) {
			slog.LogAttrs(context.TODO(), slog.LevelError, "ERROR", slog.String("err", fmt.Sprintf("Expected message-nu-%d, got %s", i, resp.Message)))
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Delete 1000 examples
	for _, id := range exampleIds {
		deleteExample(id)

		time.Sleep(10 * time.Millisecond)
	}
}
