package WebQuery

import (
	"rustQueryProxy/Config"
	"time"
)

type serverModel struct {
	Address    string `json:"addr"`
	GamePort   int    `json:"gameport"`
	SteamId    string `json:"steamid"`
	Name       string `json:"name"`
	AppId      int    `json:"appid"`
	GameDir    string `json:"gamedir"`
	Version    string `json:"version"`
	Product    string `json:"product"`
	Region     int    `json:"region"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Bots       int    `json:"bots"`
	Map        string `json:"map"`
	Secure     bool   `json:"secure"`
	Dedicated  bool   `json:"dedicated"`
	Os         string `json:"os"`
	Tags       string `json:"gametype"`
}

type serverListResponse struct {
	Response struct {
		Servers []*serverModel `json:"servers"`
	} `json:"response"`
}

type cachedDataModel struct {
	Items      map[string]*serverModel
	UpdateTime time.Time
}

func (c *cachedDataModel) IsExpired() bool {
	return c.UpdateTime.Add(Config.SteamCacheTime).Before(time.Now())
}
