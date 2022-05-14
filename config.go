package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Config struct {
	Bind                         string        // Server bind address
	IpWhitelist                  []string      // Whitelist for incoming connections
	QueryIntervalInSeconds       time.Duration // Time until next game-server update
	ServerCacheTimeInSeconds     time.Duration // Game-server cache time
	QueryConnectTimeoutInSeconds time.Duration // Game-server connection timeout when updating data
	UpdateBurstLimit             int           // Number of simultaneous updates
	CustomTagsWhitelist          []string      // Server custom tags whitelist
}

var cfg Config

func LoadConfig() {
	file, err := os.Open("config.json")
	if os.IsNotExist(err) {
		cfg = Config{
			Bind:                         "0.0.0.0:5050",
			IpWhitelist:                  []string{"127.0.0.1"},
			QueryIntervalInSeconds:       30,
			ServerCacheTimeInSeconds:     60,
			QueryConnectTimeoutInSeconds: 5,
			UpdateBurstLimit:             5,
			CustomTagsWhitelist:          []string{"monthly", "biweekly", "weekly", "vanilla", "pve", "roleplay", "creative", "softcore", "minigame", "training", "battlefield", "broyale", "builds"},
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
