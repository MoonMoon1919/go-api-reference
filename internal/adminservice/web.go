package adminservice

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/moonmoon1919/go-api-reference/internal/cache"
	"github.com/moonmoon1919/go-api-reference/internal/middleware"
	"github.com/moonmoon1919/go-api-reference/internal/requests"
	"github.com/moonmoon1919/go-api-reference/internal/responses"
	"github.com/moonmoon1919/go-api-reference/pkg/events"
)

const (
	msgNotFoundError     = "NOT_FOUND"
	msgJsonMarshallError = "JSON_MARSHAL_ERROR"
	msgServiceError      = "SERVICE_ERROR"
	msgControllerError   = "CONTROLLER_ERROR"
	keyError             = "ERROR"
	errInvalidLimit      = "LIMIT_MUST_BE_INTEGER"
	errLimitOutOfRange   = "LIMIT_OUT_OF_RANGE"
	errInvalidPage       = "PAGE_MUST_BE_INTEGER"
	errPageOutOfRange    = "PAGE_OUT_OF_RANGE"
	errMissingUser       = "USER_MISSING"
	pathValId            = "id"
)

type Controller struct {
	Service Service
	Cache   cache.Cacher
}

// MARK: User
func (c Controller) AddUser(w http.ResponseWriter, r *http.Request) {
	var request CreateUserRequest
	if err := requests.LoadRequestBody(w, r, &request); err != nil {
		return
	}

	err := c.Service.AddUser(r.Context(), request.UserId)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgServiceError,
			slog.String(keyError, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	responses.WriteCreatedResponse(w, nil, &responses.Headers{
		responses.NoCachePrivate(),
	})
}

func (c Controller) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := requests.LoadPathValue(r, pathValId)
	if err != nil {
		responses.WriteInternalServerErrorResponse(w)
		return
	}

	data, err := c.Service.GetUser(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, userServiceNotFound):
			// Not found is not necessarily an error from our perspective
			// So we log it as info
			slog.LogAttrs(
				r.Context(),
				slog.LevelInfo,
				notFoundMsg,
				slog.Any(logKeyId, id),
			)
			responses.WriteNotFoundResponse(w)
			return
		default:
			slog.LogAttrs(
				r.Context(),
				slog.LevelError,
				msgServiceError,
				slog.Any(keyError, err),
			)
			responses.WriteInternalServerErrorResponse(w)
			return
		}
	}

	resp := NewGetUserResponseFromUser(&data)
	respBytes, err := json.Marshal(resp)

	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgJsonMarshallError,
			slog.String(keyError, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	// Content-Digest
	digest := responses.CalculateContentDigest(&respBytes)

	responses.WriteSuccessResponse(
		w,
		&respBytes,
		&responses.Headers{
			responses.NoCachePrivate(),
			responses.ContentDigest(digest, responses.SHA256),
		},
	)

	return
}

func (c Controller) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, page, err := requests.GetPaginationParameters(r)
	if err != nil {
		responses.WriteBadRequestResponse(w, err.Error())
		return
	}

	data, err := c.Service.ListUsers(r.Context(), limit, page)
	if err != nil {
		switch {
		// Dont log client errors
		case errors.Is(err, limitToLargeError):
			responses.WriteBadRequestResponse(w, err.Error())
			return
		case errors.Is(err, invalidPageError):
			responses.WriteBadRequestResponse(w, err.Error())
			return
		default:
			slog.LogAttrs(
				r.Context(),
				slog.LevelError,
				msgServiceError,
				slog.String(keyError, err.Error()),
			)

			responses.WriteInternalServerErrorResponse(w)
			return
		}
	}

	resp := NewListUsersResponseFromUsers(&data)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgJsonMarshallError,
			slog.String(keyError, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
	}

	digest := responses.CalculateContentDigest(&respBytes)

	responses.WriteSuccessResponse(
		w,
		&respBytes,
		&responses.Headers{
			responses.NoCachePrivate(),
			responses.ContentDigest(digest, responses.SHA256),
		},
	)
}

func (c Controller) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if id, err := requests.LoadPathValue(r, pathValId); err != nil {
		responses.WriteInternalServerErrorResponse(w)
		return
	} else {
		err := c.Service.DeleteUser(r.Context(), id)

		if err != nil {
			switch {
			case errors.Is(err, userServiceNotFound):
				// Log client errors as info
				slog.LogAttrs(
					r.Context(),
					slog.LevelInfo,
					msgNotFoundError,
					slog.String(keyError, err.Error()),
				)

				responses.WriteNotFoundResponse(w)
				return
			default:
				slog.LogAttrs(
					r.Context(),
					slog.LevelInfo,
					msgServiceError,
					slog.String(keyError, err.Error()),
				)

				responses.WriteInternalServerErrorResponse(w)
				return
			}
		}
	}

	responses.WriteNoContentResponse(w, &responses.Headers{
		responses.NoCachePrivate(),
	})
}

// MARK: Example
func (c Controller) GetExample(w http.ResponseWriter, r *http.Request) {
	id, err := requests.LoadPathValue(r, pathValId)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgControllerError,
			slog.String(errKey, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	data, err := c.Service.GetExample(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, exampleServiceNotFound):
			// Log client errors as info
			slog.LogAttrs(
				r.Context(),
				slog.LevelInfo,
				msgNotFoundError,
				slog.String(keyError, err.Error()),
			)

			responses.WriteNotFoundResponse(w)
			return
		default:
			slog.LogAttrs(
				r.Context(),
				slog.LevelInfo,
				msgServiceError,
				slog.String(keyError, err.Error()),
			)

			responses.WriteInternalServerErrorResponse(w)
			return
		}
	}

	resp := NewGetExampleResponseFromExample(data)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgJsonMarshallError,
			slog.String(keyError, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	digest := responses.CalculateContentDigest(&respBytes)

	responses.WriteSuccessResponse(
		w,
		&respBytes,
		&responses.Headers{
			responses.NoCachePrivate(),
			responses.ContentDigest(digest, responses.SHA256),
		},
	)
}

func (c Controller) GetExamplesForUser(w http.ResponseWriter, r *http.Request) {
	userId, err := requests.LoadPathValue(r, pathValId)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgControllerError,
			slog.String(errKey, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	limit, page, err := requests.GetPaginationParameters(r)
	if err != nil {
		responses.WriteBadRequestResponse(w, err.Error())
		return
	}

	data, err := c.Service.GetExamplesForUser(r.Context(), userId, limit, page)
	if err != nil {
		switch {
		case errors.Is(err, limitToLargeError):
			responses.WriteBadRequestResponse(w, err.Error())
			return
		case errors.Is(err, invalidPageError):
			responses.WriteBadRequestResponse(w, err.Error())
			return
		default:
			slog.LogAttrs(
				r.Context(),
				slog.LevelError,
				msgServiceError,
				slog.String(keyError, err.Error()),
			)

			responses.WriteInternalServerErrorResponse(w)
			return
		}
	}

	resp := NewListExampleResponseFromExamples(data)
	respByes, err := json.Marshal(resp)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgJsonMarshallError,
			slog.String(keyError, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	digest := responses.CalculateContentDigest(&respByes)

	responses.WriteSuccessResponse(
		w,
		&respByes,
		&responses.Headers{
			responses.NoCachePrivate(),
			responses.ContentDigest(digest, responses.SHA256),
		},
	)

	return
}

func (c Controller) DeleteExample(w http.ResponseWriter, r *http.Request) {
	id, err := requests.LoadPathValue(r, pathValId)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgControllerError,
			slog.String(errKey, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgServiceError,
			slog.String(keyError, errMissingUser),
		)
		responses.WriteInternalServerErrorResponse(w)
		return
	}

	err = c.Service.DeleteExample(r.Context(), user.Id, id)
	if err != nil {
		switch {
		case errors.Is(err, exampleServiceNotFound):
			// Log client errors as info
			slog.LogAttrs(
				r.Context(),
				slog.LevelInfo,
				msgNotFoundError,
				slog.String(keyError, err.Error()),
			)

			responses.WriteNotFoundResponse(w)
			return
		default:
			slog.LogAttrs(
				r.Context(),
				slog.LevelError,
				msgServiceError,
				slog.String(keyError, err.Error()),
			)

			responses.WriteInternalServerErrorResponse(w)
			return
		}
	}

	// Clear the cache
	c.Cache.Delete(r.Context(), id)

	responses.WriteNoContentResponse(w, &responses.Headers{
		responses.NoCachePrivate(),
	})

	return
}

// MARK: Audit
func (c Controller) GetEventsForItem(w http.ResponseWriter, r *http.Request) {
	id, err := requests.LoadPathValue(r, pathValId)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgControllerError,
			slog.String(logKeyErr, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	limit, page, err := requests.GetPaginationParameters(r)
	if err != nil {
		responses.WriteBadRequestResponse(w, err.Error())
		return
	}

	data, err := c.Service.GetEventsForItem(r.Context(), id, limit, page)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgServiceError,
			slog.String(keyError, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	// Marshal response
	resp := NewListEventsFromEventsList(&data)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgJsonMarshallError,
			slog.String(logKeyErr, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	// Generate content digest
	digest := responses.CalculateContentDigest(&respBytes)

	// Write response
	responses.WriteSuccessResponse(w, &respBytes, &responses.Headers{
		responses.NoCachePrivate(),
		responses.ContentDigest(digest, responses.SHA256),
	})
}

func (c Controller) GetEventsForUser(w http.ResponseWriter, r *http.Request) {
	id, err := requests.LoadPathValue(r, pathValId)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgControllerError,
			slog.String(logKeyErr, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	limit, page, err := requests.GetPaginationParameters(r)
	if err != nil {
		responses.WriteBadRequestResponse(w, err.Error())
		return
	}

	// Check if we're filtering by event name
	eventName := r.URL.Query().Get("eventName")

	var data []events.Event
	var svcErr error

	if eventName == "" {
		data, svcErr = c.Service.GetEventsForUser(r.Context(), id, limit, page)
	} else {
		data, svcErr = c.Service.GetByEventAndUser(r.Context(), id, eventName, limit, page)
	}

	if svcErr != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgServiceError,
			slog.String(keyError, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	// Marshal response
	resp := NewListEventsFromEventsList(&data)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgJsonMarshallError,
			slog.String(logKeyErr, err.Error()),
		)
	}

	// Generate content digest
	digest := responses.CalculateContentDigest(&respBytes)

	// Write response
	responses.WriteSuccessResponse(w, &respBytes, &responses.Headers{
		responses.NoCachePrivate(),
		responses.ContentDigest(digest, responses.SHA256),
	})
}
