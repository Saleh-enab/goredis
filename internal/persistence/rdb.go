package persistence

import (
	"encoding/gob"
	"fmt"
	"log/slog"
	"os"
	"path"
	"redis/internal/config"
	"redis/internal/db"
	"time"
)

type SnapshotTracker struct {
	keys   int
	ticker time.Ticker
	rdb    *config.RDBSnapshot
}

func newSnapshotTracker(rdb *config.RDBSnapshot) *SnapshotTracker {
	return &SnapshotTracker{
		keys:   0,
		ticker: *time.NewTicker(time.Second * time.Duration(rdb.Secs)),
		rdb:    rdb,
	}
}

var trackers = []*SnapshotTracker{}

func InitRDBTracker(conf *config.Config) {
	for _, rdb := range conf.Rdb {
		tracker := newSnapshotTracker(&rdb)
		trackers = append(trackers, tracker)

		go func() {
			defer tracker.ticker.Stop()

			for range tracker.ticker.C {
				slog.Info(fmt.Sprintf("keys changed %d - keys required to change %d", tracker.keys, tracker.rdb.KeysChanged))
				if tracker.keys >= tracker.rdb.KeysChanged {
					saveRDB(conf)
				}
				tracker.keys = 0
			}
		}()
	}
}

func IncrRDBTickers() {

	for _, t := range trackers {
		t.keys++
	}
}

func saveRDB(conf *config.Config) {
	fp := path.Join(conf.Dir, conf.RdbFilename)
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("error opening the RDB file", "err", err)
		return
	}
	defer f.Close()

	err = gob.NewEncoder(f).Encode(db.Data)
	if err != nil {
		slog.Error("error saving the RDB file", "err", err)
		return
	}

	slog.Info("saved RDB file successfully")
}

func SyncRDB(conf *config.Config) {
	fp := path.Join(conf.Dir, conf.RdbFilename)

	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		slog.Error("error opening RDB file", "err", err)
		return
	}

	err = gob.NewDecoder(f).Decode(&db.Data)
	if err != nil {
		slog.Error("error decoding RDB file", "err", err)
		return
	}
}
