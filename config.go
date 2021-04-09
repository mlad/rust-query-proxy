package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Config struct {
	Address             string        `json:"address"`               // Server address
	Whitelist           []string      `json:"whitelist"`             // Whitelist for incoming connections
	ServerUpdateTime    time.Duration `json:"server_update_time"`    // Time until next game-server update
	ServerCacheTime     time.Duration `json:"server_cache_time"`     // Game-server cache time
	QueryConnectTimeout time.Duration `json:"query_connect_timeout"` // Game-server connection timeout when updating data
	UpdateBurstLimit    int           `json:"update_burst_limit"`    // Number of simultaneous updates
}

var cfg Config

func LoadConfig() {
	file, err := os.Open("config.json")
	if os.IsNotExist(err) {
		cfg = Config{
			Address:             "0.0.0.0:5050",
			Whitelist:           []string{"127.0.0.1"},
			ServerUpdateTime:    30 * time.Second,
			ServerCacheTime:     time.Minute,
			QueryConnectTimeout: 5 * time.Second,
			UpdateBurstLimit:    5,
		}
		SaveConfig()
		return
	} else if err != nil {
		log.Fatalf("LoadConfig: %s\n", err.Error())
	}

	decoder := json.NewDecoder(file)
	_ = decoder.Decode(&cfg)

	_ = file.Close()
}

func SaveConfig() {
	file, err := os.Create("config.json")
	if err != nil {
		log.Fatalf("SaveConfig: %s\n", err.Error())
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(cfg)

	_ = file.Close()
}
