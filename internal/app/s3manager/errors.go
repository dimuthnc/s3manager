package s3manager

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

// Error codes that may be returned from an S3 client.
const (
	ErrBucketDoesNotExist = "The specified bucket does not exist"
	ErrKeyDoesNotExist    = "The specified key does not exist"
)

// errInvalidMove is returned when a move/copy request is malformed. It maps to
// an HTTP 400 response.
var errInvalidMove = errors.New("invalid move request")

// errEmptyArchive is returned when a folder download would produce an empty
// archive. It maps to an HTTP 404 response.
var errEmptyArchive = errors.New("no objects to archive")

// handleHTTPError handles HTTP errors.
func handleHTTPError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError

	var se *json.SyntaxError
	ok := errors.As(err, &se)
	if ok {
		code = http.StatusUnprocessableEntity
	}

	switch {
	case errors.Is(err, errInvalidMove):
		code = http.StatusBadRequest
	case errors.Is(err, errEmptyArchive):
		code = http.StatusNotFound
	case errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF):
		code = http.StatusUnprocessableEntity
	case strings.Contains(err.Error(), ErrBucketDoesNotExist) || strings.Contains(err.Error(), ErrKeyDoesNotExist):
		code = http.StatusNotFound
	}

	http.Error(w, http.StatusText(code), code)

	// Log if server error
	if code >= http.StatusInternalServerError {
		log.Println(err)
	}
}
