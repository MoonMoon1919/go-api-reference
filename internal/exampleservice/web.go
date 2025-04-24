package exampleservice

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/moonmoon1919/go-api-reference/internal/cache"
	"github.com/moonmoon1919/go-api-reference/internal/middleware"
	"github.com/moonmoon1919/go-api-reference/internal/requests"
	"github.com/moonmoon1919/go-api-reference/internal/responses"
)

const (
	msgInternalServerError      = "INTERNAL_SERVER_ERROR"
	msgJsonMarshallError        = "JSON_MARSHAL_ERROR"
	msgServiceError             = "SERVICE_ERROR"
	msgNotFound                 = "NOT_FOUND"
	msgCacheError               = "CACHE_ERROR"
	msgControllerError          = "CONTROLLER_ERROR"
	cacheSetError               = "CACHE_SET_ERROR"
	cacheCheckGet               = "CACHE_CHECK_GET"
	cacheHit                    = "CACHE_HIT"
	cacheMiss                   = "CACHE_MISS"
	cacheCheckConditionalUpdate = "CACHE_CHECK_CONDITIONAL_UPDATE"
	cacheMissConditionalUpdate  = "CACHE_MISS_CONDITIONAL_UPDATE"
	cacheHitConditionalUpdate   = "CACHE_HIT_CONDITIONAL_UPDATE"
	errMissingUser              = "USER_MISSING"
	errInvalidLimit             = "LIMIT_MUST_BE_INTEGER"
	errLimitOutOfRange          = "LIMIT_OUT_OF_RANGE"
	errInvalidPage              = "PAGE_MUST_BE_INTEGER"
	errPageOutOfRange           = "PAGE_OUT_OF_RANGE"
	keyError                    = "ERROR"
	etagLog                     = "ETAG"
	pathValId                   = "id"
)

type Controller struct {
	Service Service
	Cache   cache.Cacher
}

// MARK: GET
func (c Controller) Get(w http.ResponseWriter, r *http.Request) {
	id, err := requests.LoadPathValue(r, pathValId)
	if err != nil {
		responses.WriteInternalServerErrorResponse(w)
		return
	}

	// Cache check
	etag := r.Header.Get(requests.IfNoneMatch.Name())
	if len(etag) != 0 {
		condition := cache.NewNoneMatch(c.Cache)

		slog.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			cacheCheckGet, slog.String(etagLog, etag),
		)

		// Cache hit
		if ok := condition.Met(r.Context(), id, etag); ok {
			slog.LogAttrs(
				r.Context(),
				slog.LevelInfo,
				cacheHit,
				slog.String(etagLog, etag),
			)

			responses.WriteNotModifiedResponse(w, &responses.Headers{
				responses.NoCachePrivate(),
			})

			return
		} else {
			slog.LogAttrs(
				r.Context(),
				slog.LevelInfo,
				cacheMiss,
				slog.String(etagLog, etag),
			)
		}
	}

	// Call the service
	data, err := c.Service.Get(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, repositoryNotFoundError):
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

	// Marshal the response in the controller
	// So we don't inadvertently call json.Marshall a bunch of times
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

	// Read thru cache
	cacheKey, err := c.Cache.Set(r.Context(), id, &respBytes)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgCacheError,
			slog.String(keyError, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	// Content-Digest
	digest := responses.CalculateContentDigest(&respBytes)

	// Write the response
	responses.WriteSuccessResponse(
		w,
		&respBytes,
		&responses.Headers{
			responses.NoCachePrivate(),
			responses.Etag(cacheKey),
			responses.ContentDigest(digest, responses.SHA256),
		},
	)
}

// MARK: LIST
func (c Controller) List(w http.ResponseWriter, r *http.Request) {
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

	// Limit for pagination
	queryLimit := r.URL.Query().Get("limit")

	var limit int = 10 // Default limit is 10
	if queryLimit != "" {
		n, err := strconv.Atoi(queryLimit)

		if err != nil {
			responses.WriteBadRequestResponse(w, errInvalidLimit)
			return
		}

		// Smallest limit we support is 1
		// Sending a limit less than 1 results in an error
		if n < 1 {
			responses.WriteBadRequestResponse(w, errLimitOutOfRange)
			return
		}

		limit = n
	}

	// Page for pagination
	queryPage := r.URL.Query().Get("page")

	var page int = 1
	if queryPage != "" {
		n, err := strconv.Atoi(queryPage)

		if err != nil {
			responses.WriteBadRequestResponse(w, errInvalidPage)
			return
		}

		// Pages are 1-based
		// Sending in a page number less than 1 results in an error
		if n < 1 {
			responses.WriteBadRequestResponse(w, errPageOutOfRange)
			return
		}

		page = n
	}

	// Call the service
	data, err := c.Service.List(r.Context(), user.Id, limit, page)
	if err != nil {
		switch {
		// Client error - dont log as an error
		case errors.Is(err, limitToLargeError):
			responses.WriteBadRequestResponse(w, err.Error())
			return
		// Client error - dont log as an error
		case errors.Is(err, RepositoryListError):
			responses.WriteInternalServerErrorResponse(w)
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

	// Write the response
	responses.WriteSuccessResponse(
		w,
		&respBytes,
		&responses.Headers{
			responses.NoCachePrivate(),
			responses.ContentDigest(digest, responses.SHA256),
		},
	)

}

// MARK: CREATE (POST)
func (c Controller) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgServiceError,
			slog.Any(keyError, errMissingUser),
		)
		responses.WriteInternalServerErrorResponse(w)
		return
	}

	// Parse the request
	var request CreateExampleRequest
	if err := requests.LoadRequestBody(w, r, &request); err != nil {
		return
	}

	// Call the service
	data, err := c.Service.Add(r.Context(), user.Id, request.Message)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgServiceError,
			slog.String(keyError, err.Error()),
		)

		switch {
		case errors.Is(err, InvalidMessageError{}):
			responses.WriteBadRequestResponse(w, err.Error())
			return
		case errors.Is(err, repositoryAddError):
			responses.WriteInternalServerErrorResponse(w)
			return
		default:
			responses.WriteInternalServerErrorResponse(w)
			return
		}
	}

	// Create the response bytes
	resp := NewCreateExampleResponseFromExample(data)
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

	// Write thru cache
	cacheKey, err := c.Cache.Set(r.Context(), data.Id, &respBytes)
	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			cacheSetError,
			slog.Any(keyError, err),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	// Content-Digest
	digest := responses.CalculateContentDigest(&respBytes)

	// Write response
	responses.WriteCreatedResponse(
		w,
		&respBytes,
		&responses.Headers{
			responses.NoCachePrivate(),
			responses.Etag(cacheKey),
			responses.ContentDigest(digest, responses.SHA256),
		},
	)
}

// MARK: PATCH
func (c Controller) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := requests.LoadPathValue(r, pathValId)
	if err != nil {
		responses.WriteInternalServerErrorResponse(w)
		return
	}

	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgServiceError,
			slog.Any(keyError, errMissingUser),
		)
		responses.WriteInternalServerErrorResponse(w)
		return
	}

	// Cache condition
	etag := r.Header.Get(requests.IfMatch.Name())

	if len(etag) != 0 {
		condition := cache.NewMatch(c.Cache)

		slog.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			cacheCheckConditionalUpdate,
			slog.String(etagLog, etag),
		)

		// Cache miss - precondition failed
		if ok := condition.Met(r.Context(), id, etag); !ok {
			slog.LogAttrs(
				r.Context(),
				slog.LevelInfo,
				cacheMissConditionalUpdate,
				slog.String(etagLog, etag),
			)

			responses.WritePreflightConditionFailedResponse(w)
			return
		} else {
			slog.LogAttrs(
				r.Context(),
				slog.LevelInfo,
				cacheHitConditionalUpdate,
				slog.String(etagLog, etag),
			)
		}
	}

	var request PatchExampleRequest
	if err := requests.LoadRequestBody(w, r, &request); err != nil {
		return
	}

	data, err := c.Service.Update(r.Context(), user.Id, id, request.Message)

	if err != nil {

		switch {
		case errors.Is(err, repositoryNotFoundError):
			// Log client errors as info
			slog.LogAttrs(
				r.Context(),
				slog.LevelInfo,
				msgNotFound,
				slog.String(keyError, err.Error()),
			)

			responses.WriteNotFoundResponse(w)
			return
		case errors.Is(err, InvalidMessageError{}):
			responses.WriteBadRequestResponse(w, err.Error())
			return
		case errors.Is(err, repositoryAddError):
			slog.LogAttrs(
				r.Context(),
				slog.LevelError,
				msgServiceError,
				slog.String(keyError, err.Error()),
			)

			responses.WriteInternalServerErrorResponse(w)
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

	// Create response
	resp := NewPatchExampleResponseFromExample(data)
	respBytes, err := json.Marshal(resp)

	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgInternalServerError,
			slog.String(keyError, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	}

	// Write thru cache
	cacheKey, err := c.Cache.Set(r.Context(), id, &respBytes)

	if err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			cacheSetError,
			slog.Any(keyError, err),
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
			responses.Etag(cacheKey),
			responses.ContentDigest(digest, responses.SHA256),
		},
	)
}

// MARK: DELETE
func (c Controller) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgServiceError,
			slog.Any(keyError, errMissingUser),
		)
		responses.WriteInternalServerErrorResponse(w)
		return
	}

	if id, err := requests.LoadPathValue(r, pathValId); err != nil {
		slog.LogAttrs(
			r.Context(),
			slog.LevelError,
			msgControllerError,
			slog.String(errKey, err.Error()),
		)

		responses.WriteInternalServerErrorResponse(w)
		return
	} else {
		err := c.Service.Delete(r.Context(), user.Id, id)

		if err != nil {
			switch {
			case errors.Is(err, repositoryNotFoundError):
				// Log client errors as info
				slog.LogAttrs(
					r.Context(),
					slog.LevelInfo,
					msgNotFound,
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
	}

	responses.WriteNoContentResponse(w, &responses.Headers{
		responses.NoCachePrivate(),
	})
}
