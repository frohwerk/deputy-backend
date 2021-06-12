package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/frohwerk/deputy-backend/internal/database"
)

type store struct {
	*sql.DB
}

func main() {
	db := &store{database.Open()}
	// rows, err := db.Query(`SELECT id, name FROM draft.apps WHERE apps.id = $1`, "demo")
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "%v", err)
	// 	os.Exit(1)
	// }
	rows, err := db.Query(`SELECT id, name FROM draft.vapps_components WHERE apps.id = $1`, "demo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	// Read apps (exactly one)
	// Read apps_components_all for app
	// Read deployments_all for all components
	// Create timeline by creating a snapshot for each distinct timestamp (valid_from & valid_until)
	// --> matching valid_until & valid_from should only create two snapshots, not three
	// apps_components: membership can only begin or end => add or remove component
	// deployments: image_ref can change, deployments can be missing (not deployed at that point in time)
}

type app struct {
	name    string
	history []snapshot
}

type snapshot struct {
	time       time.Time
	components []component
}

type component struct {
	name  string
	image string
}

func (db *store) method() {

}
