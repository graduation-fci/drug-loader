package model

type Interaction struct {
	Name   string `json:"name"`
	Level  string `json:"level"`
	Effect string `json:"effect"`
}

type Drug struct {
	Name         string        `json:"name"`
	Url          string        `json:"url,omitempty"`
	Interactions []Interaction `json:"interactions"`
}
