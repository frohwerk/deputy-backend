package main

import (
	"database/sql"
	"strings"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/frohwerk/deputy-backend/pkg/api"

	apps "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/watch"
)

var Log logger.Logger = logger.Default

type deploymentStore interface {
	SetImage(componentId, platformId, imageRef string) (*database.Deployment, error)
	RemoveImage(componentId, platformId string) error
}

type componentStore interface {
	CreateIfAbsent(name string) (*database.Component, error)
	GetByName(name string) (*database.Component, error)
}

type deploymentHandler struct {
	platform    *api.Platform
	components  componentStore
	deployments deploymentStore
}

func (h *deploymentHandler) handleEvent(event watch.Event) error {
	depl, ok := event.Object.(*apps.Deployment)
	if !ok {
		return nil
	}

	switch event.Type {
	case watch.Added:
		fallthrough
	case watch.Modified:
		return h.registerDeployment(depl)
	case watch.Deleted:
		return h.unregisterDeployment(depl)
	default:
		Log.Warn("Unexpected event type: %s", event.Type)
		return nil
	}
}

func (h *deploymentHandler) registerDeployment(depl *apps.Deployment) error {
	c, err := h.components.CreateIfAbsent(depl.Name)
	if err != nil {
		return err
	}
	log.Debug("Component '%s' is registered with id '%s'", depl.Name, c.Id)

	d, err := h.deployments.SetImage(c.Id, h.platform.Id, strings.TrimPrefix(depl.Spec.Template.Spec.Containers[0].Image, "docker-pullable://"))
	switch {
	case err != nil:
		return err
	case d == nil:
		log.Trace("Component '%s' is unmodified", depl.Name)
	default:
		log.Debug("Updated image for component %s to %s\n", c.Name, d.ImageRef)
	}

	return nil
}

func (h *deploymentHandler) unregisterDeployment(depl *apps.Deployment) error {
	c, err := h.components.GetByName(depl.Name)
	switch {
	case err == sql.ErrNoRows:
		return nil // Nothing to do in this case...
	case err != nil:
		return err
	}
	log.Debug("Removing deployment of component '%s'", c.Id)
	return h.deployments.RemoveImage(c.Id, h.platform.Id)
}
