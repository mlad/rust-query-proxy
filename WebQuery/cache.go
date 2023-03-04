package WebQuery

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

const cacheFileName = "steam_cache.json"

var cachedData *cachedDataModel

func LoadCache() {
	file, err := os.Open(cacheFileName)

	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Fatalf("[WebQuery.tryLoadCache] Failed to open file: %s\n", err.Error())
	}

	cachedData = new(cachedDataModel)

	decoder := json.NewDecoder(file)
	err = decoder.Decode(cachedData)
	if err != nil {
		log.Fatalf("[WebQuery.tryLoadCache] Failed to deserialize json: %s\n", err.Error())
	}

	_ = file.Close()
}

func updateCache(servers map[string]*serverModel) {
	file, err := os.Create(cacheFileName)
	if err != nil {
		log.Fatalf("[WebQuery.saveCache] Failed to create file: %s\n", err.Error())
	}

	cachedData = &cachedDataModel{
		Items:      servers,
		UpdateTime: time.Now(),
	}

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cachedData)
	if err != nil {
		log.Fatalf("[WebQuery.saveCache] Failed to serialize json: %s\n", err.Error())
	}

	_ = file.Close()
}
