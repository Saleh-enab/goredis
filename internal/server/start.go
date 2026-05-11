package server

import (
	"log/slog"
	"net"

	"redis/internal/app"
	"redis/internal/config"
)

func Start(addr string) error {
	slog.Info("Reading the config file...")
	conf := config.ReadConf("./redis.conf")
	state := app.NewAppState(conf)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	slog.Info("server is listening", "addr", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error("accept failed", "err", err)
			continue
		}

		go HandleConnection(conn, state)
	}
}
