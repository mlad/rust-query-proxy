package WebQuery

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"rustQueryProxy/Config"
	"rustQueryProxy/RustServer"
	"strings"
	"sync"
)

var lock sync.Mutex

func getServers() (map[string]*serverModel, error) {
	if cachedData != nil && !cachedData.IsExpired() {
		return cachedData.Items, nil
	}

	lock.Lock()
	defer lock.Unlock()

	if cachedData != nil && !cachedData.IsExpired() {
		return cachedData.Items, nil
	}

	client := http.Client{}
	uri := fmt.Sprintf("https://api.steampowered.com/IGameServersService/GetServerList/v1/?key=%s&filter=gamedir%%5Crust&limit=9999999", Config.SteamApiToken)
	response, err := client.Get(uri)
	if err != nil {
		log.Fatalf("[WebQuery.getServers] Failed to make HTTP request: %s\n", err.Error())
	}

	var payload serverListResponse

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		log.Fatalf("[WebQuery.getServers] Failed to deserialize json: %s\n", err.Error())
	}

	result := make(map[string]*serverModel)
	for _, v := range payload.Response.Servers {
		addressParts := strings.Split(v.Address, ":")
		result[fmt.Sprintf("%s:%d", addressParts[0], v.GamePort)] = v
	}

	updateCache(result)
	return result, nil
}

func Query(address string) (*RustServer.RawModel, error) {
	if Config.SteamApiToken == "" {
		log.Fatalf("[WebQuery.Query] SteamApiToken must be specified")
	}
	if Config.SteamCacheTime <= 0 {
		log.Fatalf("[WebQuery.Query] SteamCacheTime must be positive number")
	}

	allServers, err := getServers()
	if err != nil {
		return nil, err
	}

	serverData, ok := allServers[address]
	if !ok {
		return nil, errors.New("server address is not found in the cache")
	}

	result := &RustServer.RawModel{
		Hostname: serverData.Name,
		Map:      serverData.Map,
		Tags:     strings.Split(serverData.Tags, ","),
	}

	return result, nil
}
