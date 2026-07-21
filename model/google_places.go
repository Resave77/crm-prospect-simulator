package model

type TextSearchRequest struct {
	TextQuery string `json:"textQuery"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
