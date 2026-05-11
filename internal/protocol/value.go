package protocol

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
)

type ValueType string

const (
	Array   ValueType = "*"
	Bulk    ValueType = "$"
	String  ValueType = "+"
	Integer ValueType = ":"
	Error   ValueType = "-"
	Null    ValueType = ""
)

type Value struct {
	Type    ValueType
	Bulk    string
	String  string
	Integer int
	Error   string
	Array   []Value
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}

func (v *Value) ReadArray(r *bufio.Reader) error {
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

func Deserialize(v *Value) []byte {
	switch v.Type {

	case Array:
		var result []byte

		// array header
		result = append(result, []byte(fmt.Sprintf("*%d\r\n", len(v.Array)))...)

		// append each item
		for _, item := range v.Array {
			result = append(result, Deserialize(&item)...)
		}

		return result

	case String:
		return []byte(fmt.Sprintf("+%s\r\n", v.String))

	case Integer:
		return []byte(fmt.Sprintf(":%d\r\n", v.Integer))

	case Bulk:
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v.Bulk), v.Bulk))

	case Error:
		return []byte(fmt.Sprintf("-%s\r\n", v.Error))

	case Null:
		return []byte("$-1\r\n")
	default:
		slog.Error("Invalid type received")
		return nil
	}
}
