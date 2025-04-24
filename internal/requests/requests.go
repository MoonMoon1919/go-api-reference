package requests

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/moonmoon1919/go-api-reference/internal/responses"
)

type HeaderKey string

func (h HeaderKey) Name() string {
	return string(h)
}

const (
	IfNoneMatch                 HeaderKey = "If-None-Match"
	IfMatch                     HeaderKey = "If-Match"
	msgMissingRequestBody                 = "MISSING_REQUEST_BODY"
	msgInvalidRequestBody                 = "INVALID_REQUEST_BODY"
	msgMissingExpectedPathParam           = "MISSING_EXPECTED_PATH_PARAM"
	keyError                              = "ERROR"
	keyPath                               = "PATH"
	keyParam                              = "PARAM"
)

var LimitTooLowError = errors.New("limit must be greater than 0")
var InvalidLimitError = errors.New("limit must be an integer")
var PageTooLowError = errors.New("page must be greater than 0")
var InvalidPageError = errors.New("page must be an integer")

func LoadRequestBody(w http.ResponseWriter, r *http.Request, v any) error {
	if r.Body == nil {
		responses.WriteBadRequestResponse(w, msgMissingRequestBody)
		return errors.New(msgMissingRequestBody)
	}

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		slog.Error(msgInvalidRequestBody, keyError, err)

		if err.Error() == "EOF" {
			responses.WriteBadRequestResponse(w, msgMissingRequestBody)
		} else if err.Error() == "unexpected EOF" {
			responses.WriteBadRequestResponse(w, msgInvalidRequestBody)
		} else {
			responses.WriteBadRequestResponse(w, err.Error())
		}

		return err
	}

	return nil
}

func LoadPathValue(r *http.Request, key string) (string, error) {
	value := r.PathValue(key)

	if value == "" {
		slog.Error(msgMissingExpectedPathParam, keyPath, r.URL.Path, keyParam, key)

		return "", errors.New(msgMissingExpectedPathParam)
	}

	return value, nil
}

func GetPaginationParameters(r *http.Request) (int, int, error) {
	queryLimit := r.URL.Query().Get("limit")
	var limit int = 10
	if queryLimit != "" {
		n, err := strconv.Atoi(queryLimit)

		if err != nil {
			return 0, 0, InvalidLimitError
		}

		if n < 1 {
			return 0, 0, LimitTooLowError
		}

		limit = n
	}

	queryPage := r.URL.Query().Get("page")
	var page int = 1
	if queryPage != "" {
		n, err := strconv.Atoi(queryPage)

		if err != nil {
			return 0, 0, InvalidPageError
		}

		if n < 1 {
			return 0, 0, PageTooLowError
		}

		page = n
	}

	return limit, page, nil
}
