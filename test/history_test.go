package test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestStuff(t *testing.T) {
	db := DB()
	err := Clean(db)
	if assert.NoError(t, err) {
		t.Log("INSERTs ok")
	} else {
		t.Log("INSERTs failed:", err)
	}
}

func Clean(db *sql.DB) error {
	statements := []string{
		`DELETE FROM envs`,
		`DELETE FROM platforms`,
		`DELETE FROM apps`,
		`DELETE FROM components`,
		`DELETE FROM deployments`,
		`DELETE FROM apps_timeline`,
		`DELETE FROM apps_components`,
		`DELETE FROM apps_components_history`,
		`DELETE FROM deployments_history`,

		`INSERT INTO envs (id, name) VALUES ('example', 'Example')`,
		`INSERT INTO platforms (id, env_id, name, api_server, namespace, secret) VALUES ('minishift', 'example', 'Minishift', 'https://192.168.178.31:8443', 'my-namespace', '')`,

		`INSERT INTO envs (id, name) VALUES ('integration', 'Environment for user acceptance testing')`,
		`INSERT INTO platforms (id, env_id, name, api_server, namespace, secret) VALUES ('minishift-si', 'integration', 'Minishift (SI)', 'https://192.168.178.31:8443', 'my-namespace', '')`,

		`INSERT INTO apps (id, name) VALUES ('tester', 'Test-Anwendung')`,
		`INSERT INTO components (id, name) VALUES ('component-a', 'Irgendeine Komponente')`,
		`INSERT INTO components (id, name) VALUES ('component-b', 'Eine andere Komponente')`,

		`INSERT INTO deployments (component_id, platform_id, image_ref) VALUES ('component-a', 'minishift', 'image-registry.cluster.local/my-namespace/a:1.0.2')`,
		`INSERT INTO deployments (component_id, platform_id, image_ref) VALUES ('component-b', 'minishift', 'image-registry.cluster.local/my-namespace/b:4.1')`,
		`INSERT INTO deployments (component_id, platform_id, image_ref) VALUES ('component-a', 'minishift-si', 'image-registry.cluster.local/my-namespace/a:1.0.2')`,
	}

	for _, statement := range statements {
		_, err := db.Exec(statement)
		if err != nil {
			fmt.Printf("statement execution failed:\n%s\n%s\n", statement, err)
			return err
		}
	}

	return nil
}
