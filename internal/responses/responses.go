package responses

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type DigestAlgorithm string

const (
	SHA256 DigestAlgorithm = "sha-256"
)

type HeaderKey string

func (h HeaderKey) Name() string {
	return string(h)
}

type HeaderValue string

func (h HeaderValue) Value() string {
	return string(h)
}

const (
	ContentType         HeaderKey   = "Content-Type"
	ContentDigestKey    HeaderKey   = "Content-Digest"
	CacheControlKey     HeaderKey   = "Cache-Control"
	EtagKey             HeaderKey   = "Etag"
	NoCacheValue        HeaderValue = "no-cache"
	NoCachePrivateValue HeaderValue = "no-cache, private"
	ApplicationJson     HeaderValue = "application/json"
)

// HTTP Error Messages
const (
	UNAUTHORIZED          = "UNAUTHORIZED"
	NOT_FOUND             = "NOT_FOUND"
	CONFLICT              = "CONFLICT"
	PRECONDITION_FAILED   = "PRECONDITION_FAILED"
	INTERNAL_SERVER_ERROR = "INTERNAL_SERVER_ERROR"
)

const HTTP_ERROR_DEFAULT = `{"error": "INTERNAL_SERVER_ERROR"}`

type Header struct {
	key   HeaderKey
	value string
}

func (h Header) Key() string {
	return h.key.Name()
}

func (h Header) Value() string {
	return h.value
}

type Headers []Header

// MARK: Content Digest
func CalculateContentDigest(val *[]byte) string {
	hash := sha256.Sum256(*val)
	encodedHash := base64.StdEncoding.EncodeToString(hash[:])

	return encodedHash
}

// MARK: Cache Condition
func NoCache() Header {
	return Header{key: CacheControlKey, value: NoCacheValue.Value()}
}

func NoCachePrivate() Header {
	return Header{key: CacheControlKey, value: NoCachePrivateValue.Value()}
}

func ContentDigest(data string, alg DigestAlgorithm) Header {
	return Header{key: ContentDigestKey, value: fmt.Sprintf("%s=%s", alg, data)}
}

func Etag(value string) Header {
	return Header{key: EtagKey, value: value}
}

// MARK: Responses
type errorResponse struct {
	Error string `json:"error"`
}

func writeResponse(w http.ResponseWriter, data *[]byte, code int, headers *Headers) {
	w.Header().Set(ContentType.Name(), ApplicationJson.Value())
	for _, header := range *headers {
		w.Header().Set(header.Key(), header.Value())
	}
	w.WriteHeader(code)

	if data == nil {
		return
	}

	// Dereference the pointer to data here
	// WHY?
	// We have intermediate convenience functions between controllers
	// and this function and we don't want to make many copies of the response payload
	// This is negligible for small payloads but for large payloads
	// Can make a difference in how much memory we're consuming
	// Since this function doesnt know, we err on the side of performance first
	w.Write(*data)
}

func writeErrorResponse(w http.ResponseWriter, error string, code int) {
	e := errorResponse{
		Error: error,
	}

	respBytes, err := json.Marshal(e)

	if err != nil {
		http.Error(w, HTTP_ERROR_DEFAULT, http.StatusInternalServerError)
		return
	}

	writeResponse(w, &respBytes, code, &Headers{})
}

// MARK: Error Responses
func WriteBadRequestResponse(w http.ResponseWriter, error string) {
	writeErrorResponse(w, error, http.StatusBadRequest)
}

func WriteUnauthorizedResponse(w http.ResponseWriter) {
	writeErrorResponse(w, UNAUTHORIZED, http.StatusUnauthorized)
}

func WriteNotFoundResponse(w http.ResponseWriter) {
	writeErrorResponse(w, NOT_FOUND, http.StatusNotFound)
}

func WriteConflictResponse(w http.ResponseWriter) {
	writeErrorResponse(w, CONFLICT, http.StatusConflict)
}

func WritePreflightConditionFailedResponse(w http.ResponseWriter) {
	writeErrorResponse(w, PRECONDITION_FAILED, http.StatusPreconditionFailed)
}

func WriteInternalServerErrorResponse(w http.ResponseWriter) {
	writeErrorResponse(w, INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
}

// MARK: Success Responses
func WriteSuccessResponse(w http.ResponseWriter, data *[]byte, headers *Headers) {
	writeResponse(w, data, http.StatusOK, headers)
}

func WriteAcceptedResponse(w http.ResponseWriter, data *[]byte, headers *Headers) {
	writeResponse(w, data, http.StatusAccepted, headers)
}

func WriteCreatedResponse(w http.ResponseWriter, data *[]byte, headers *Headers) {
	writeResponse(w, data, http.StatusCreated, headers)
}

func WriteNoContentResponse(w http.ResponseWriter, headers *Headers) {
	writeResponse(w, nil, http.StatusNoContent, headers)
}

func WriteNotModifiedResponse(w http.ResponseWriter, headers *Headers) {
	writeResponse(w, nil, http.StatusNotModified, headers)
}
