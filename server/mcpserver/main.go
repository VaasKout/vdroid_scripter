package main

import (
	"context"
	"log"
	"os"
)

func main() {
	baseURL := os.Getenv("VDROID_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8081"
	}

	server := New(baseURL)
	if err := server.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
