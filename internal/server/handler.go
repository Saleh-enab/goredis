package server

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"

	"redis/internal/commands"
	"redis/internal/protocol"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		v := protocol.Value{Type: protocol.Array}

		if err := v.ReadArray(reader); err != nil {
			if err == io.EOF {
				slog.Info("client disconnected")
				return
			}
			slog.Error("read error", "err", err)
			return
		}

		handleCommand(writer, &v)

		fmt.Println(v.Array)
	}
}

func handleCommand(w *bufio.Writer, v *protocol.Value) {
	cmd := strings.ToUpper(v.Array[0].Bulk)

	handler, ok := commands.Handlers[cmd]
	if !ok {
		sendResponse(w, &protocol.Value{Type: protocol.Error, Error: "ERR unknown command '" + cmd + "'"})
		return
	}

	res := handler(v)
	sendResponse(w, res)
}

func sendResponse(w *bufio.Writer, v *protocol.Value) {
	switch v.Type {

	case protocol.Array:
		w.Write([]byte(fmt.Sprintf("*%d\r\n", len(v.Array))))

		for _, item := range v.Array {
			sendResponse(w, &item)
		}
		return

	case protocol.String:
		w.Write([]byte(fmt.Sprintf("+%s\r\n", v.String)))

	case protocol.Bulk:
		w.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v.Bulk), v.Bulk)))

	case protocol.Error:
		w.Write([]byte(fmt.Sprintf("-%s\r\n", v.Error)))

	case protocol.Null:
		w.Write([]byte("$-1\r\n"))
	}

	w.Flush()
}
