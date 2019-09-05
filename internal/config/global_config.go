package config

import (
	"github.com/BurntSushi/toml"
)

type ClientConfig struct {
	Paths []string `toml:"paths"`
	/* AllowFlags       []string `toml:"allow_flags"`
	AllowCompression string   `toml:"allow_compression"`
	RequireFlags     []string `toml:"require_flags"` */
}

type RuntimeConfig struct {
	SocketWrapper string `toml:"socket_wrapper"`
	SocketDir     string `toml:"socket_dir"`
	Home          string
}

type GlobalConfig struct {
	Clients map[string]ClientConfig `toml:"client"`
	Runtime RuntimeConfig
}

func LoadGlobalConfig(path string) (*GlobalConfig, error) {
	result := &GlobalConfig{}
	_, err := toml.DecodeFile(path, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
