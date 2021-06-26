package handler

import (
	"github.com/frohwerk/deputy-backend/cmd/imgmatch/matcher"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/logger"
)

var Log logger.Logger = logger.Default

type handler struct {
	matcher.Matcher
	database.ImageLinker
}

func New(m matcher.Matcher, l database.ImageLinker) *handler {
	return &handler{m, l}
}

func (h *handler) Accept(image string) {
	Log.Info("Inspecting image '%s'", image)
	if c, err := h.ImageLinker.Count(image); err != nil {
		Log.Error("error checking existing assignment for image '%s': %s", image, err)
	} else if c > 0 {
		Log.Debug("image '%s' has already an assigned artifact. skipping", image)
	}
	artifacts, err := h.Matcher.Match(image)
	if err != nil {
		Log.Error("error matching image %s: %s", image, err)
	}
	if len(artifacts) == 0 {
		Log.Warn("No matching artifact found for image '%s'", image)
	}
	for _, file := range artifacts {
		h.ImageLinker.AddLink(image, file.Id)
	}
}
