package formjson

type SVCRForm struct {
	ServerURL string `json:"ServerURL"`
	Username  string `json:"Username"`
	Password  string `json:"Password"`
}

type Voucher struct {
	ServerURL         string `json:"ServerURL"`
	Secret            string `json:"Secret"`
	AuthType          string `json:"AuthType"`
	ThingsPanelApiKey string `json:"ThingsPanelApiKey"`
	ThingsPanelApiURL string `json:"ThingsPanelApiURL"`
}
