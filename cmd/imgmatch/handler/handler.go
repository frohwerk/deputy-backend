package handler

import (
	"log"
	"net/http"

	"github.com/frohwerk/deputy-backend/cmd/imgmatch/matcher"
	"github.com/frohwerk/deputy-backend/internal/database"
)

type handler struct {
	matcher.Matcher
	database.ImageLinker
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	image := r.FormValue("image")
	switch {
	case r.Method != http.MethodPost:
		methodNotAllowed(rw)
	case image == "":
		invalidParameter(rw)
	default:
		go h.Accept(image)
		rw.WriteHeader(http.StatusAccepted)
		rw.Write(nil)
	}
}

func New(m matcher.Matcher, l database.ImageLinker) http.Handler {
	return &handler{m, l}
}

func (h *handler) Accept(image string) {
	artifacts, err := h.Matcher.Match(image)
	if err != nil {
		log.Printf("error matching image %s: %s\n", image, err)
	}
	for _, file := range artifacts {
		h.ImageLinker.AddLink(image, file.Id)
	}
}

func methodNotAllowed(rw http.ResponseWriter) {
	rw.Header().Add("Allow", "POST")
	rw.WriteHeader(http.StatusMethodNotAllowed)
	rw.Write(nil)
}

func invalidParameter(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write(nil)
}
