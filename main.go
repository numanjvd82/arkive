package main

import (
	"log"

	"arkive/internal/router"
)

func main() {
	r := router.New()

	if err := r.Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
