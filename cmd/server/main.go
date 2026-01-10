package main

import (
	"log"
	"propertyProject/internal"
)

func main() {
	cfg := internal.LoadConfig()
	server := internal.NewServer(cfg)

	log.Printf("Starting server on %s\n", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
