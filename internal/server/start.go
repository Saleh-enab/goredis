package server

import (
	"log/slog"
	"net"
)

func Start(addr string) error {
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

		go HandleConnection(conn)
	}
}
