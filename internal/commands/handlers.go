package commands

import (
	"log/slog"

	"redis/internal/app"
	"redis/internal/protocol"
)

type Handler func(*protocol.Value, *app.AppState) *protocol.Value

var Handlers = map[string]Handler{
	"GET":     Get,
	"SET":     Set,
	"PING":    ping,
	"COMMAND": command,
}

func Get(v *protocol.Value, state *app.AppState) *protocol.Value {
	args := v.Array[1:]

	if len(args) != 1 {
		return &protocol.Value{Type: protocol.Error, Error: "ERR Invalid number of arguments for 'GET' command"}
	}

	key := args[0].Bulk

	val, ok := app.Data.Get(key)
	if !ok {
		return &protocol.Value{Type: protocol.Null}
	}

	return &protocol.Value{Type: protocol.Bulk, Bulk: val}
}

func Set(v *protocol.Value, state *app.AppState) *protocol.Value {
	args := v.Array[1:]

	if len(args) != 2 {
		return &protocol.Value{Type: protocol.Error, Error: "ERR Invalid number of arguments for 'SET' command"}
	}

	key := args[0].Bulk
	val := args[1].Bulk

	app.Data.Set(key, val)

	if state.Conf.AofEnabled && state.Aof != nil && state.Aof.W != nil {
		slog.Info("saving AOF record")
		state.Aof.W.Write(protocol.Deserialize(v))
		if state.Conf.AofFsync == "always" {
			state.Aof.W.Flush()
		}
	}

	return &protocol.Value{Type: protocol.String, String: "OK"}
}

func ping(_ *protocol.Value, state *app.AppState) *protocol.Value {
	return &protocol.Value{Type: protocol.String, String: "PONG"}
}

func command(_ *protocol.Value, state *app.AppState) *protocol.Value {
	return &protocol.Value{
		Type:  protocol.Array,
		Array: []protocol.Value{},
	}
}
