package commands

import (
	"fmt"

	"redis/internal/protocol"
)

var DB = map[string]string{}

type Handler func(*protocol.Value) *protocol.Value

var Handlers = map[string]Handler{
	"GET":  get,
	"SET":  set,
	"PING": ping,
}

func get(v *protocol.Value) *protocol.Value {
	args := v.Array[1:]

	if len(args) != 1 {
		return &protocol.Value{Type: protocol.Error, Error: "ERR Invalid number of arguments for 'GET' command"}
	}

	key := args[0].Bulk

	val, ok := DB[key]
	if !ok {
		return &protocol.Value{Type: protocol.Null}
	}

	return &protocol.Value{Type: protocol.Bulk, Bulk: val}
}

func set(v *protocol.Value) *protocol.Value {
	args := v.Array[1:]

	if len(args) != 2 {
		return &protocol.Value{Type: protocol.Error, Error: "ERR Invalid number of arguments for 'SET' command"}
	}

	key := args[0].Bulk
	val := args[1].Bulk

	DB[key] = val
	fmt.Println(DB)

	return &protocol.Value{Type: protocol.String, String: "OK"}
}

func ping(_ *protocol.Value) *protocol.Value {
	return &protocol.Value{Type: protocol.String, String: "PONG"}
}

func command(_ *protocol.Value) *protocol.Value {
	return &protocol.Value{
		Type:  protocol.Array,
		Array: []protocol.Value{},
	}
}
