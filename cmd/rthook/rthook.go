package main

import (
	"fmt"
	"os"

	artifactory "github.com/frohwerk/deputy-backend/internal/artifactory/client"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/spf13/cobra"
)

var (
	command = &cobra.Command{Run: Run}

	rtbase string
	port   int
)

func init() {
	command.Flags().StringVarP(&rtbase, "artifactory", "r", "http://localhost:8091/libs-release-local", "base-uri for the artifactory server")
	command.Flags().IntVarP(&port, "port", "p", 8082, "port this webhook will listen on")
}

func main() {
	command.Use = os.Args[0]
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	}
}

func Run(cmd *cobra.Command, args []string) {
	db := database.Open()
	defer db.Close()

	rt := artifactory.New(rtbase)
	eh := &EventHandler{Repository: rt, FileCreater: database.NewFileStore(db)}
	rt.OnArtifactDeployed(eh.OnArtifactDeployed)

	server := &server{port: port, handler: rt.WebhookHandler}

	server.start()
}
