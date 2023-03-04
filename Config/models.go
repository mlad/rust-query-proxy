package Config

type appConfigPayload struct {
	Bind                         string
	IpWhitelist                  []string
	QueryIntervalInSeconds       int
	ServerCacheTimeInSeconds     int
	QueryConnectTimeoutInSeconds int
	UpdateBurstLimit             int
	CustomTagsWhitelist          []string
	SteamApiToken                string
	SteamCacheTimeInSeconds      int
}
