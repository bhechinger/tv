package config

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"runtime"
)

type Config struct {
	Database DB
	RSSFeed RSS
}

type RSS struct {
	BaseURL string
	Key string
}

type DB struct {
	User     string
	Host     string
	Port     string
	Password string
	DBName   string
	SSLMode  string
}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func Get(filename string) (Config, error) {
	var config Config

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	// TODO: Figure out how best to set defaults
	if _, err := toml.Decode(string(data), &config); err != nil {
		return config, err
	}
	return config, nil
}
