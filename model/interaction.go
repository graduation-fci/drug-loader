package model

type Interaction struct {
	Name               string `json:"name"`
	HashedName         string `json:"hashedName"`
	Level              string `json:"level"`
	ConsumerEffect     string `json:"consumerEffect"`
	ProfessionalEffect string `json:"professionalEffect"`
}

type Drug struct {
	Name         string        `json:"name"`
	Url          string        `json:"url,omitempty"`
	Interactions []Interaction `json:"interactions"`
}
