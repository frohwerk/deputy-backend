package envs

import (
	"fmt"
	"net/http"
)

type statusError struct {
	error
	statusCode int
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func badRequest(m string, values ...interface{}) statusError {
	return statusError{fmt.Errorf(m, values...), http.StatusBadRequest}
}

func notFound(m string, values ...interface{}) statusError {
	return statusError{fmt.Errorf(m, values...), http.StatusNotFound}
}

func writeErrorResponse(rw http.ResponseWriter, err error) {
	switch err := err.(type) {
	case statusError:
		rw.WriteHeader(err.statusCode)
	default:
		rw.WriteHeader(http.StatusInternalServerError)
	}
	writeJsonResponse(rw, ErrorResponse{fmt.Sprintf("%s", err)})
}
