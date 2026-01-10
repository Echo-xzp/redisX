package main

import (
	"log"
	"redisx/internal/server"
)

func main() {
	s := server.NewServer(":6379")
	if err := s.Start(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
