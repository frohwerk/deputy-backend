package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var err error
var db *sql.DB

func TestMain(m *testing.M) {
	db, err = sql.Open("postgres", "postgres://test:drowssap@database:5432/test?sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

loop:
	for i := 0; i < 15; i++ {
		fmt.Println("Trying to connect to database postgres://database:5432/test?sslmode=disable")
		err = db.Ping()
		if err != nil {
			fmt.Println("Failed to connect to database. Sleeping...")
			time.Sleep(time.Second)
		} else {
			fmt.Println("Connected to database postgres://database:5432/test?sslmode=disable")
			break loop
		}
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rc := m.Run()

	os.Exit(rc)
}

func TestStuff(t *testing.T) {
	err := CleanDatabase()
	if assert.NoError(t, err) {
		t.Log("INSERTs ok")
	} else {
		t.Log("INSERTs failed:", err)
	}
}

func CleanDatabase() error {
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
		fmt.Println(statement)
		_, err := db.Exec(statement)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}
