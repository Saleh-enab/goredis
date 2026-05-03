package persistence

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path"

	"redis/internal/config"
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
