package Config

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

const appConfigFileName = "config.json"

var (
	Bind                string        // Server bind address
	IpWhitelist         []string      // Whitelist for incoming connections
	QueryInterval       time.Duration // Time until next game-server update
	ServerCacheTime     time.Duration // Game-server cache time
	QueryConnectTimeout time.Duration // Game-server connection timeout when updating data
	UpdateBurstLimit    int           // Number of simultaneous updates
	CustomTagsWhitelist []string      // Server custom tags whitelist
	SteamApiToken       string        // Steam API token
	SteamCacheTime      time.Duration // Steam Web API server list cache time
)

func Load() {
	file, err := os.Open(appConfigFileName)

	if os.IsNotExist(err) {
		Bind = "0.0.0.0:5050"
		IpWhitelist = []string{"127.0.0.1"}
		QueryInterval = 30 * time.Second
		ServerCacheTime = 60 * time.Second
		QueryConnectTimeout = 5 * time.Second
		UpdateBurstLimit = 5
		CustomTagsWhitelist = []string{
			"monthly", "biweekly", "weekly",
			"vanilla", "hardcore", "softcore",
			"pve", "roleplay", "creative", "minigame", "training", "battlefield", "broyale", "builds",
			"NA", "SA", "EU", "WA", "EA", "OC", "AF"}
		SteamApiToken = ""
		SteamCacheTime = 5 * time.Minute
		Save()
		return
	} else if err != nil {
		log.Fatalf("[Config.Load] Failed to open file: %s\n", err.Error())
	}

	var payload appConfigPayload

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&payload)
	if err != nil {
		log.Fatalf("[Config.Load] Failed to deserialize json: %s\n", err.Error())
	}

	_ = file.Close()

	Bind = payload.Bind
	IpWhitelist = payload.IpWhitelist
	QueryInterval = time.Duration(payload.QueryIntervalInSeconds) * time.Second
	ServerCacheTime = time.Duration(payload.ServerCacheTimeInSeconds) * time.Second
	QueryConnectTimeout = time.Duration(payload.QueryConnectTimeoutInSeconds) * time.Second
	UpdateBurstLimit = payload.UpdateBurstLimit
	CustomTagsWhitelist = payload.CustomTagsWhitelist
	SteamApiToken = payload.SteamApiToken
	SteamCacheTime = time.Duration(payload.SteamCacheTimeInSeconds) * time.Second
}

func Save() {
	file, err := os.Create(appConfigFileName)
	if err != nil {
		log.Fatalf("[Config.Save] Failed to create file: %s\n", err.Error())
	}

	payload := appConfigPayload{
		Bind:                         Bind,
		IpWhitelist:                  IpWhitelist,
		QueryIntervalInSeconds:       int(QueryInterval / time.Second),
		ServerCacheTimeInSeconds:     int(ServerCacheTime / time.Second),
		QueryConnectTimeoutInSeconds: int(QueryConnectTimeout / time.Second),
		UpdateBurstLimit:             UpdateBurstLimit,
		CustomTagsWhitelist:          CustomTagsWhitelist,
		SteamApiToken:                SteamApiToken,
		SteamCacheTimeInSeconds:      int(SteamCacheTime / time.Second),
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(payload)
	if err != nil {
		log.Fatalf("[Config.Save] Failed to serialize json: %s\n", err.Error())
	}

	_ = file.Close()
}
