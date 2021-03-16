package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	stuff string

	workshop = &cobra.Command{Run: run}
)

func init() {
	workshop.PersistentFlags().StringVarP(&stuff, "name", "n", "World", "Your name. Or any name. Just hello world things")
}

func main() {
	workshop.Use = os.Args[0]
	if err := workshop.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	fmt.Printf("Hallo %s!", stuff)
}
