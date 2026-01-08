package main

import (
	"log"

	"claude-watcher/internal/app"
)

func main() {
	cfg, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
