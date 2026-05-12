package app

import (
	"time"

	"redis/internal/config"
	"redis/internal/persistence"
)

type AppState struct {
	Conf *config.Config
	Aof  *persistence.Aof
	RDB  *persistence.RDBState
}

func NewAppState(conf *config.Config) *AppState {
	if conf.AofEnabled {
		persistence.ReplayAOF(conf)
	}

	if len(conf.Rdb) > 0 {
		persistence.SyncRDB(conf)
	}

	state := AppState{
		Conf: conf,
		RDB:  &persistence.RDBState{},
	}

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

	if len(conf.Rdb) > 0 {
		persistence.InitRDBTracker(conf, state.RDB)
	}

	return &state
}
