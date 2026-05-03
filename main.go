package main

import (
	"log"
	"log/slog"

	"redis/internal/server"
)

func main() {
	slog.Info("Reading teh config file...")
	readConf("./redis.conf")

	const addr = ":6379"

	if err := server.Start(addr); err != nil {
		log.Fatalf("cannot listen on port %s: %v", addr, err)
	}
}
