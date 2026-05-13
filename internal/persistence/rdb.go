package persistence

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
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

func InitRDBTracker(conf *config.Config, state *RDBState) {
	for _, rdb := range conf.Rdb {
		tracker := newSnapshotTracker(&rdb)
		trackers = append(trackers, tracker)

		go func() {
			defer tracker.ticker.Stop()

			for range tracker.ticker.C {
				// slog.Info(fmt.Sprintf("keys changed %d - keys required to change %d", tracker.keys, tracker.rdb.KeysChanged))
				if tracker.keys >= tracker.rdb.KeysChanged {
					SaveRDB(conf, state)
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

func SaveRDB(conf *config.Config, state *RDBState) {
	// Build the full path of the RDB snapshot file.
	fp := path.Join(conf.Dir, conf.RdbFilename)

	// Open the file in read/write mode and truncate old contents.
	// O_RDWR is required because we later read the file again
	// to verify its checksum.
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		slog.Error("RDB - error opening the RDB file", "err", err)
		return
	}
	defer f.Close()

	var buf bytes.Buffer

	slog.Info("RDB - saving DB to RDB file")

	// Encode the database state into an in-memory buffer.
	//
	// During BGSAVE we serialize the copied snapshot to avoid
	// locking the main DB while writes are happening.
	//
	// Otherwise, we lock the DB for reading and serialize
	// the current live map directly.
	if state.BGSaveRunning {
		err = gob.NewEncoder(&buf).Encode(&state.DBCopy)
	} else {
		db.Data.Mu.RLock()
		err = gob.NewEncoder(&buf).Encode(&db.Data.M)
		db.Data.Mu.RUnlock()
	}

	if err != nil {
		slog.Error("RDB - error encoding DB", "err", err)
		return
	}

	data := buf.Bytes()

	// Compute the checksum of the encoded buffer before writing.
	// This is later compared against the file checksum to verify
	// that the written data matches the original buffer.
	bsum, err := Hash(bytes.NewReader(data))
	if err != nil {
		slog.Error("RDB - cannot compute buffer checksum", "err", err)
		return
	}

	// Write the serialized snapshot to disk.
	_, err = f.Write(data)
	if err != nil {
		slog.Error("RDB - cannot write to the RDB file", "err", err)
		return
	}

	// Force the OS to flush buffered data to stable storage.
	// Without Sync, the data may still only exist in the OS cache.
	err = f.Sync()
	if err != nil {
		slog.Error("RDB - cannot sync RDB file", "err", err)
		return
	}

	// Reset the file cursor to the beginning before reading it again.
	_, err = f.Seek(0, 0)
	if err != nil {
		slog.Error("RDB - cannot seek file", "err", err)
		return
	}

	// Compute the checksum of the persisted file.
	fusm, err := Hash(f)
	if err != nil {
		slog.Error("RDB - cannot compute file checksum", "err", err)
		return
	}

	// Verify that the file contents exactly match the original buffer.
	if bsum != fusm {
		slog.Error(fmt.Sprintf(
			"RDB - buffer and file checksums do not match: buffer= %s\n file= %s\n",
			bsum,
			fusm,
		))
		return
	}

	slog.Info("RDB - saved RDB file successfully")
}

func SyncRDB(conf *config.Config) {
	fp := path.Join(conf.Dir, conf.RdbFilename)

	f, err := os.Open(fp)
	if err != nil {
		slog.Error("error opening RDB file", "err", err)
		f.Close()
		return
	}

	err = gob.NewDecoder(f).Decode(&db.Data.M)
	if err != nil {
		slog.Error("error decoding RDB file", "err", err)
		return
	}
}

func Hash(r io.Reader) (string, error) {
	h := sha256.New()

	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
