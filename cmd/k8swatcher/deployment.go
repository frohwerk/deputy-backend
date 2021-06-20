package main

import (
	"strings"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"

	apps "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/watch"
)

type deploymentStore interface {
	SetImage(componentId, platformId, imageRef string) (*database.Deployment, error)
}

type componentStore interface {
	CreateIfAbsent(name string) (*database.Component, error)
}

type deploymentHandler struct {
	platform    *api.Platform
	components  componentStore
	deployments deploymentStore
}

func (h *deploymentHandler) handleEvent(event watch.Event) error {
	obj, ok := event.Object.(*apps.Deployment)
	if !ok {
		return nil
	}

	c, err := h.components.CreateIfAbsent(obj.Name)
	if err != nil {
		return err
	}
	log.Debug("Component '%s' is registered with id '%s'", obj.Name, c.Id)

	d, err := h.deployments.SetImage(c.Id, h.platform.Id, strings.TrimPrefix(obj.Spec.Template.Spec.Containers[0].Image, "docker-pullable://"))
	switch {
	case err != nil:
		return err
	case d == nil:
		log.Trace("Component '%s' is unmodified", obj.Name)
	default:
		log.Debug("Updated image for component %s to %s\n", c.Name, d.ImageRef)
	}

	return nil
}
