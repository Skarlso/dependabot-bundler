package main

import (
	"log"

	"github.com/Skarlso/dependabot-bundler/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
