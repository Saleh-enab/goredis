package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	dir         string
	rdb         []RDBSnapshot
	rdbFilename string
	aofEnabled  bool
	aofFilename string
	aofFsync    FsyncMode
}

func NewConfig() *Config {
	return &Config{}
}

type RDBSnapshot struct {
	Secs        int
	KeysChanged int
}

type FsyncMode string

const (
	Always   FsyncMode = "always"
	EverySec FsyncMode = "everysec"
	No       FsyncMode = "no"
)

func readConf(filename string) *Config {
	conf := NewConfig()

	f, err := os.Open(filename)
	if err != nil {
		slog.Error("Error reading config file - using default config\n", "filename", filename)
		return conf
	}
	defer f.Close()

	s := bufio.NewScanner(f)

	for s.Scan() {
		l := s.Text()
		ParseLine(l, conf)
	}

	if err := s.Err(); err != nil {
		slog.Error("Error scanning the config file", "filename", filename)
		return conf
	}

	if conf.dir != "" {
		os.MkdirAll(conf.dir, 0755)
	}

	return conf
}

func ParseLine(l string, conf *Config) {
	args := strings.Split(l, " ")
	cmd := args[0]

	switch cmd {
	case "dir":
		conf.dir = args[1]

	case "appendonly":
		conf.aofEnabled = args[1] == "yes"

	case "appendfilename":
		conf.aofFilename = args[1]

	case "appenfsync":
		conf.aofFsync = FsyncMode(args[1])

	case "save":
		secs, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Invalid number of seconds")
			return
		}

		keysChanged, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Invalid number of kays")
			return
		}

		snapshot := RDBSnapshot{
			Secs:        secs,
			KeysChanged: keysChanged,
		}

		conf.rdb = append(conf.rdb, snapshot)

	case "dbfilename":
		conf.rdbFilename = args[1]
	}
}
