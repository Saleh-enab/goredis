# Redis Clone

This project is a simple Redis-compatible server written in Go. It listens on TCP port `6379`, reads configuration from `redis.conf`, and stores data in memory with support for persistence.

## Features

- Basic key-value commands: `GET`, `SET`, `DEL`, `EXISTS`, `KEYS`
- Server commands: `PING`, `COMMAND`, `DBSIZE`, `FLUSHDB`
- Persistence support:
    - AOF logging
    - RDB snapshots
- Optional password protection with `AUTH`

## Configuration

The server reads `redis.conf` at startup. The sample config in this repository enables:

- Data directory: `./data`
- AOF file: `backup.aof`
- RDB file: `backup.rdb`
- Password: `dolphins`

## Run

Start the server with:

```bash
go run .
```

The server will listen on `:6379`.

## Project Layout

- `main.go` starts the server
- `internal/server` handles connections and command dispatch
- `internal/commands` contains Redis command handlers
- `internal/persistence` manages AOF and RDB storage
- `internal/config` loads `redis.conf`

## Notes

This is a learning project and does not implement the full Redis protocol or the full Redis command set.
