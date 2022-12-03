package model

type Interaction struct {
	Name   string `json:"name"`
	Level  string `json:"level"`
	Effect string `json:"effect"`
}

type ActiveIngredient struct {
	Name         string        `json:"name"`
	Interactions []Interaction `json:"interactions"`
}
