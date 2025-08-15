package main

import (
	"log"

	"github.com/goinginblind/l0-task/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to setup application: %v", err)
	}

	application.Run()
}
