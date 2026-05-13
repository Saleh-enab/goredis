package persistence

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"

	"redis/internal/config"
	"redis/internal/db"
	"redis/internal/protocol"
)

type Aof struct {
	W    *bufio.Writer
	F    *os.File
	Conf *config.Config
}

func NewAof(conf *config.Config) *Aof {
	aof := Aof{Conf: conf}

	filePath := path.Join(conf.Dir, conf.AofFilename)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		slog.Error(fmt.Sprintf("cannot open %s", filePath), "filepath", filePath)
		return &aof
	}

	aof.W = bufio.NewWriter(f)
	aof.F = f

	return &aof
}

func ReplayAOF(conf *config.Config) {
	if conf.Dir == "" || conf.AofFilename == "" {
		return
	}

	filePath := path.Join(conf.Dir, conf.AofFilename)

	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}

		slog.Error("cannot open AOF for replay",
			"filepath", filePath,
			"err", err,
		)
		return
	}
	defer f.Close()

	r := bufio.NewReader(f)

	for {
		v := protocol.Value{}

		err := v.ReadArray(r)

		if err == io.EOF {
			break
		}

		if err != nil {
			slog.Error("unexpected error while reading AOF records", "err", err)
			break
		}

		if len(v.Array) == 0 {
			continue
		}

		cmd := strings.ToUpper(v.Array[0].Bulk)

		switch cmd {

		case "SET":
			if len(v.Array) < 3 {
				continue
			}

			db.Data.Set(
				v.Array[1].Bulk,
				v.Array[2].Bulk,
			)

		case "DEL":
			if len(v.Array) < 2 {
				continue
			}

			keys := make([]string, 0, len(v.Array)-1)

			for _, key := range v.Array[1:] {
				keys = append(keys, key.Bulk)
			}

			db.Data.Delete(keys)

		case "FLUSHDB":
			db.Data.Flush()
		}

	}
}
