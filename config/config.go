package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ApiID      int32   `json:"api_id"`
	ApiHash    string  `json:"api_hash"`
	ChannelSrc int64   `json:"channel_src"`
	ChannelDst int64   `json:"channel_dst"`
	Admin     []int64 `json:"admin"`
	OnStart bool `json:"onStart"`
    AutoFetch bool `json:"autoStart"`

	path string `json:"-"`
}

func NewConfig(path string) (*Config, error) {
	cfg, err := LoadConfig(path)
	if err == nil {
		cfg.path = path
		return cfg, nil
	}
	return &Config{path: path}, err
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	cfg.path = path
	return &cfg, nil
}

func (cfg *Config) Save() error {
	file, err := os.Create(cfg.path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(cfg)
}