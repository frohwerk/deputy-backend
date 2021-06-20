package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/frohwerk/deputy-backend/internal/util"
)

type Deployment struct {
	ComponentId string
	PlatformId  string
	ImageRef    string
	Updated     time.Time
}

type DeploymentUpdater interface {
	SetImage(componentId, platformId, imageRef string) (*Deployment, error)
}

type DeploymentLister interface {
	ListForEnv(componentId, envId string) ([]Deployment, error)
}

type DeploymentStore interface {
	DeploymentLister
	DeploymentUpdater
}

type deploymentStore struct {
	*sql.DB
}

func NewDeploymentStore(db *sql.DB) DeploymentStore {
	return &deploymentStore{db}
}

func (ds *deploymentStore) ListForEnv(componentId, envId string) ([]Deployment, error) {
	return ds.queryDeployments(`
		SELECT d.component_id, d.platform_id, d.image_ref, d.updated
		FROM platforms p
		JOIN deployments d ON d.platform_id = p.id
		WHERE component_id = $1 AND env_id = $2
	`, componentId, envId)
}

func (ds *deploymentStore) SetImage(componentId, platformId, imageRef string) (*Deployment, error) {
	return ds.queryDeployment(`
		INSERT INTO deployments (component_id, platform_id, image_ref, updated) VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (component_id, platform_id) DO
		UPDATE SET image_ref = EXCLUDED.image_ref, updated = EXCLUDED.updated
		WHERE deployments.component_id = EXCLUDED.component_id AND deployments.platform_id = EXCLUDED.platform_id AND deployments.image_ref != EXCLUDED.image_ref
		RETURNING component_id, platform_id, image_ref, updated
	`, componentId, platformId, imageRef)
}

func (ds *deploymentStore) queryDeployment(query string, args ...interface{}) (*Deployment, error) {
	return scanDeployment(ds.DB.QueryRow(query, args...))
}

func (ds *deploymentStore) queryDeployments(query string, args ...interface{}) ([]Deployment, error) {
	rows, err := ds.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)

	result := make([]Deployment, 0)
	for rows.Next() {
		if r, err := scanDeployment(rows); err != nil {
			return nil, err
		} else {
			result = append(result, *r)
		}
	}

	return result, nil
}

func scanDeployment(s scanner) (*Deployment, error) {
	entity := Deployment{}
	switch err := s.Scan(&entity.ComponentId, &entity.PlatformId, &entity.ImageRef, &entity.Updated); err {
	case nil:
		return &entity, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}
