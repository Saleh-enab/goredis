package config

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Dir         string
	Rdb         []RDBSnapshot
	RdbFilename string
	AofEnabled  bool
	AofFilename string
	AofFsync    FsyncMode
	RequirePass bool
	Password    string
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

func ReadConf(filename string) *Config {
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
		parseLine(l, conf)
	}

	if err := s.Err(); err != nil {
		slog.Error("Error scanning the config file", "filename", filename)
		return conf
	}

	if conf.Dir != "" {
		os.MkdirAll(conf.Dir, 0755)
	}

	return conf
}

func parseLine(l string, conf *Config) {
	args := strings.Split(l, " ")
	cmd := args[0]

	switch cmd {
	case "dir":
		conf.Dir = args[1]

	case "appendonly":
		conf.AofEnabled = args[1] == "yes"

	case "appendfilename":
		conf.AofFilename = args[1]

	case "appenfsync":
		conf.AofFsync = FsyncMode(args[1])

	case "save":
		secs, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Invalid number of seconds")
			return
		}

		keysChanged, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("Invalid number of keys")
			return
		}

		snapshot := RDBSnapshot{
			Secs:        secs,
			KeysChanged: keysChanged,
		}

		conf.Rdb = append(conf.Rdb, snapshot)

	case "dbfilename":
		conf.RdbFilename = args[1]

	case "requirepass":
		conf.RequirePass = true
		conf.Password = args[1]
	}
}
