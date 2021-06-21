package httputil

import (
	"fmt"
	"net/http"
)

type StatusError struct {
	error
	StatusCode int
}

type ErrorResponse struct {
	Id      *string `json:"id,omitempty"`
	Message string  `json:"message"`
}

func BadRequest(m string, values ...interface{}) StatusError {
	return StatusError{fmt.Errorf(m, values...), http.StatusBadRequest}
}

func NotFound(m string, values ...interface{}) StatusError {
	return StatusError{fmt.Errorf(m, values...), http.StatusNotFound}
}

func WriteErrorResponse(rw http.ResponseWriter, err error) {
	switch err := err.(type) {
	case StatusError:
		rw.WriteHeader(err.StatusCode)
	default:
		rw.WriteHeader(http.StatusInternalServerError)
	}
	WriteJsonResponse(rw, ErrorResponse{Message: fmt.Sprintf("%s", err)})
}
