package server

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"slices"
	"strings"

	"redis/internal/app"
	"redis/internal/client"
	"redis/internal/commands"
	"redis/internal/protocol"
)

func HandleConnection(conn net.Conn, state *app.AppState) {
	defer conn.Close()

	c := client.NewClient(conn)

	reader := bufio.NewReader(c.Conn)
	writer := bufio.NewWriter(c.Conn)

	for {
		v := protocol.Value{Type: protocol.Array}

		if err := v.ReadArray(reader); err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "forcibly closed by the remote host") {
				slog.Info("client disconnected")
				return
			}
			slog.Error("read error", "err", err)
			return
		}

		fmt.Println("received: ", v.Array)

		handleCommand(c, writer, &v, state)
	}
}

func handleCommand(c *client.Client, w *bufio.Writer, v *protocol.Value, state *app.AppState) {
	cmd := strings.ToUpper(v.Array[0].Bulk)

	if state.Conf.RequirePass && !c.Authenticated && !slices.Contains(commands.SafeCMDs, cmd) {
		sendResponse(w, &protocol.Value{Type: protocol.Error, Error: "NOAUTH authentication required"})
		return
	}

	handler, ok := commands.Handlers[cmd]
	if !ok {
		sendResponse(w, &protocol.Value{Type: protocol.Error, Error: "ERR unknown command '" + cmd + "'"})
		return
	}

	res := handler(c, v, state)
	sendResponse(w, res)
}

func sendResponse(w *bufio.Writer, v *protocol.Value) {
	w.Write(protocol.Deserialize(v))
	w.Flush()
}
