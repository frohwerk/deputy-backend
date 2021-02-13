package apps

import (
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func writeErrorResponse(resp http.ResponseWriter, err error) {
	resp.WriteHeader(http.StatusInternalServerError)
	writeJsonResponse(resp, ErrorResponse{fmt.Sprintf("%s", err)})
}
