package main

import (
	"log"

	"github.com/Skarlso/dependabot-bundler/cmd"
)

func main() {
	root := cmd.CreateRootCommand()
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
