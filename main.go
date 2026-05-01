package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"strconv"
	"strings"
)

type ValueType string

const (
	Array  ValueType = "*"
	Bulk   ValueType = "$"
	String ValueType = "+"
)

type Value struct {
	Type  ValueType
	Bulk  string
	Str   string
	Array []Value
}

// --- Helpers ---

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}

// --- Parsing ---

func (v *Value) readArray(r *bufio.Reader) error {
	line, err := readLine(r)
	if err != nil {
		return err
	}

	if len(line) == 0 || line[0] != '*' {
		return fmt.Errorf("expected array, got: %q", line)
	}

	length, err := strconv.Atoi(line[1:])
	if err != nil {
		return fmt.Errorf("invalid array length: %w", err)
	}

	v.Array = make([]Value, 0, length)

	for i := 0; i < length; i++ {
		val, err := readBulk(r)
		if err != nil {
			return err
		}
		v.Array = append(v.Array, val)
	}

	return nil
}

func readBulk(r *bufio.Reader) (Value, error) {
	line, err := readLine(r)
	if err != nil {
		return Value{}, err
	}

	if len(line) == 0 || line[0] != '$' {
		return Value{}, fmt.Errorf("expected bulk string, got: %q", line)
	}

	length, err := strconv.Atoi(line[1:])
	if err != nil {
		return Value{}, fmt.Errorf("invalid bulk length: %w", err)
	}

	data := make([]byte, length+2) // include \r\n
	if _, err := io.ReadFull(r, data); err != nil {
		return Value{}, err
	}

	return Value{
		Type: Bulk,
		Bulk: string(data[:length]),
	}, nil
}

// --- Connection Handler ---

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		v := Value{Type: Array}

		if err := v.readArray(reader); err != nil {
			if err == io.EOF {
				slog.Info("client disconnected")
				return
			}
			slog.Error("read error", "err", err)
			return
		}

		fmt.Println(v.Array)

		if _, err := conn.Write([]byte("+OK\r\n")); err != nil {
			slog.Error("write error", "err", err)
			return
		}
	}
}

// --- Main ---

func main() {
	const addr = ":6379"

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cannot listen on port %s: %v", addr, err)
	}
	defer l.Close()

	slog.Info("server is listening", "addr", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error("accept failed", "err", err)
			continue
		}

		go handleConnection(conn)
	}
}
