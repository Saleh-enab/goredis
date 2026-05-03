package app

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"redis/internal/config"
	"redis/internal/persistence"
	"redis/internal/protocol"
)

type AppState struct {
	Conf *config.Config
	Aof  *persistence.Aof
}

func NewAppState(conf *config.Config) *AppState {
	if conf.AofEnabled {
		replayAOF(conf)
	}

	state := AppState{Conf: conf}

	if conf.AofEnabled {
		state.Aof = persistence.NewAof(conf)

		if conf.AofFsync == "everysec" {
			go func() {
				t := time.NewTicker(time.Second)
				defer t.Stop()

				for range t.C {
					state.Aof.W.Flush()
				}
			}()
		}
	}

	return &state
}

func replayAOF(conf *config.Config) {
	if conf.Dir == "" || conf.AofFilename == "" {
		return
	}

	filePath := path.Join(conf.Dir, conf.AofFilename)
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		slog.Error("cannot open AOF for replay", "filepath", filePath, "err", err)
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
		if len(v.Array) < 3 || strings.ToUpper(v.Array[0].Bulk) != "SET" {
			continue
		}
		Data.Set(v.Array[1].Bulk, v.Array[2].Bulk)
	}
}
