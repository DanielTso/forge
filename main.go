// Command forge scaffolds new projects using industry-standard layouts.
package main

import (
	"os"

	"github.com/DanielTso/forge/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
