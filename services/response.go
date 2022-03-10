package services

type ProxyResponse struct {
	Ip        string  `json:"ip"`
	Port      int64   `json:"port"`
	Type      string  `json:"type"`
	Country   string  `json:"country"`
	Code      string  `json:"code"`
	City      string  `json:"city"`
	Anonymity string  `json:"anonymity"`
	Timeout   float64 `json:"timeout"`
}
