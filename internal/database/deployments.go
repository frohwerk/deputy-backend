package database

import (
	"database/sql"
	"time"
)

type Deployment struct {
	ComponentId string
	PlatformId  string
	ImageRef    string
	Updated     time.Time
}

type deploymentStore struct {
	*sql.DB
}

func NewDeploymentStore(db *sql.DB) *deploymentStore {
	return &deploymentStore{db}
}

func (ds *deploymentStore) SetImage(componentId, platformId, imageRef string) (*Deployment, error) {
	return ds.queryOne(`
		INSERT INTO deployments (component_id, platform_id, image_ref) VALUES ($1, $2, $3)
		ON CONFLICT (component_id, platform_id) DO
		UPDATE SET image_ref = EXCLUDED.image_ref, updated = CURRENT_TIMESTAMP
		RETURNING component_id, platform_id, image_ref, updated
	`, componentId, platformId, imageRef)
}

func (ds *deploymentStore) queryOne(query string, args ...interface{}) (*Deployment, error) {
	return scanDeployment(ds.DB.QueryRow(query, args...))
}

func scanDeployment(s scanner) (*Deployment, error) {
	entity := Deployment{}
	if err := s.Scan(&entity.ComponentId, &entity.PlatformId, &entity.ImageRef, &entity.Updated); err != nil {
		return nil, err
	}
	return &entity, nil
}
