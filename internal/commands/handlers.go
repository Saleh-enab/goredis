package commands

import (
	"log/slog"

	"redis/internal/app"
	"redis/internal/db"
	"redis/internal/persistence"
	"redis/internal/protocol"
)

type Handler func(*protocol.Value, *app.AppState) *protocol.Value

var Handlers = map[string]Handler{
	"GET":     Get,
	"SET":     Set,
	"DEL":     Delete,
	"EXISTS":  Exists,
	"KEYS":    Keys,
	"SAVE":    Save,
	"PING":    ping,
	"COMMAND": command,
}

func Get(v *protocol.Value, state *app.AppState) *protocol.Value {
	args := v.Array[1:]

	if len(args) != 1 {
		return &protocol.Value{Type: protocol.Error, Error: "ERR Invalid number of arguments for 'GET' command"}
	}

	key := args[0].Bulk

	val, ok := db.Data.Get(key)
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

	if len(state.Conf.Rdb) > 0 {
		persistence.IncrRDBTickers()
	}

	db.Data.Set(key, val)

	if state.Conf.AofEnabled && state.Aof != nil && state.Aof.W != nil {
		slog.Info("saving AOF record")
		state.Aof.W.Write(protocol.Deserialize(v))
		if state.Conf.AofFsync == "always" {
			state.Aof.W.Flush()
		}
	}

	return &protocol.Value{Type: protocol.String, String: "OK"}
}

func Delete(v *protocol.Value, state *app.AppState) *protocol.Value {
	args := v.Array[1:]
	var keys []string

	if len(args) < 1 {
		return &protocol.Value{Type: protocol.Error, Error: "ERR Invalid number of arguments for 'DELETE' command"}
	}

	for _, arg := range args {
		keys = append(keys, arg.Bulk)
	}

	n := db.Data.Delete(keys)

	if state.Conf.AofEnabled && state.Aof != nil && state.Aof.W != nil {
		slog.Info("saving AOF record")
		state.Aof.W.Write(protocol.Deserialize(v))
		if state.Conf.AofFsync == "always" {
			state.Aof.W.Flush()
		}
	}

	return &protocol.Value{Type: protocol.Integer, Integer: n}
}

func Exists(v *protocol.Value, state *app.AppState) *protocol.Value {
	args := v.Array[1:]
	var keys []string

	if len(args) < 1 {
		return &protocol.Value{Type: protocol.Error, Error: "ERR Invalid number of arguments for 'EXISTS' command"}
	}

	for _, arg := range args {
		keys = append(keys, arg.Bulk)
	}

	n := db.Data.Exists(keys)
	return &protocol.Value{Type: protocol.Integer, Integer: n}
}

func Keys(v *protocol.Value, state *app.AppState) *protocol.Value {
	args := v.Array[1:]
	if len(args) > 1 {
		return &protocol.Value{Type: protocol.Error, Error: "ERR Invalid number of arguments for 'KEYS' command"}
	}

	pattern := args[0].Bulk
	matches := db.Data.Keys(pattern)

	replay := protocol.Value{Type: protocol.Array}
	for _, m := range matches {
		replay.Array = append(replay.Array, protocol.Value{Type: protocol.Bulk, Bulk: m})
	}

	return &replay
}

func Save(v *protocol.Value, state *app.AppState) *protocol.Value {
	persistence.SaveRDB(state.Conf)
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
