package main

import (
	"log"

	"github.com/Skarlso/dependabot-bundler-action/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
