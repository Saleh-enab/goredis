package main

import (
	"log"

	"redis/internal/server"
)

func main() {

	const addr = ":6379"

	if err := server.Start(addr); err != nil {
		log.Fatalf("cannot listen on port %s: %v", addr, err)
	}
}
