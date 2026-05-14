package commands

import (
	"log/slog"
	"maps"

	"redis/internal/app"
	"redis/internal/client"
	"redis/internal/db"
	"redis/internal/persistence"
	"redis/internal/protocol"
)

type Handler func(*client.Client, *protocol.Value, *app.AppState) *protocol.Value

var Handlers = map[string]Handler{
	"GET":     Get,
	"SET":     Set,
	"DEL":     Delete,
	"EXISTS":  Exists,
	"KEYS":    Keys,
	"SAVE":    Save,
	"BGSAVE":  BGSave,
	"DBSIZE":  DBSize,
	"FLUSHDB": FlushDB,
	"AUTH":    Auth,
	"PING":    ping,
	"COMMAND": command,
}

var SafeCMDs = []string{
	"COMMAND", "AUTH",
}

func Get(c *client.Client, v *protocol.Value, state *app.AppState) *protocol.Value {
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

func Set(c *client.Client, v *protocol.Value, state *app.AppState) *protocol.Value {
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

func Delete(c *client.Client, v *protocol.Value, state *app.AppState) *protocol.Value {
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

func Exists(c *client.Client, v *protocol.Value, state *app.AppState) *protocol.Value {
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

func Keys(c *client.Client, v *protocol.Value, state *app.AppState) *protocol.Value {
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

func Save(c *client.Client, v *protocol.Value, state *app.AppState) *protocol.Value {
	persistence.SaveRDB(state.Conf, state.RDB)
	return &protocol.Value{Type: protocol.String, String: "OK"}
}

func BGSave(c *client.Client, v *protocol.Value, state *app.AppState) *protocol.Value {
	if state.RDB.BGSaveRunning {
		return &protocol.Value{Type: protocol.Error, Error: "ERR background saving already in progress"}
	}

	cp := make(map[string]string, len(db.Data.M))

	db.Data.Mu.RLock()
	maps.Copy(cp, db.Data.M)
	db.Data.Mu.RUnlock()

	state.RDB.BGSaveRunning = true
	state.RDB.DBCopy = cp

	go func() {
		defer func() {
			state.RDB.BGSaveRunning = false
			state.RDB.DBCopy = nil
		}()

		persistence.SaveRDB(state.Conf, state.RDB)
	}()

	return &protocol.Value{Type: protocol.String, String: "OK"}
}

func DBSize(c *client.Client, v *protocol.Value, state *app.AppState) *protocol.Value {
	db.Data.Mu.RLock()
	size := len(db.Data.M)
	db.Data.Mu.RUnlock()

	return &protocol.Value{Type: protocol.Integer, Integer: size}
}

func FlushDB(c *client.Client, v *protocol.Value, state *app.AppState) *protocol.Value {
	db.Data.Flush()

	if state.Conf.AofEnabled && state.Aof != nil && state.Aof.W != nil {
		slog.Info("saving AOF record")
		state.Aof.W.Write(protocol.Deserialize(v))
		if state.Conf.AofFsync == "always" {
			state.Aof.W.Flush()
		}
	}

	go persistence.SaveRDB(state.Conf, state.RDB)

	return &protocol.Value{Type: protocol.String, String: "OK"}
}

func Auth(c *client.Client, v *protocol.Value, state *app.AppState) *protocol.Value {
	args := v.Array[1:]

	if len(args) != 1 {
		return &protocol.Value{Type: protocol.Error, Error: "ERR Invalid number of arguments for 'AUTH' command"}
	}

	pass := args[0].Bulk
	if pass == state.Conf.Password {
		c.Authenticated = true
		return &protocol.Value{Type: protocol.String, String: "OK"}
	} else {
		c.Authenticated = false
		return &protocol.Value{Type: protocol.Error, Error: "ERR invalid password"}
	}
}

func ping(c *client.Client, _ *protocol.Value, state *app.AppState) *protocol.Value {
	return &protocol.Value{Type: protocol.String, String: "PONG"}
}

func command(c *client.Client, _ *protocol.Value, state *app.AppState) *protocol.Value {
	return &protocol.Value{
		Type:  protocol.Array,
		Array: []protocol.Value{},
	}
}
