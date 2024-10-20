package models

// Plan is a model for a pricing plan.
type Plan struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Features    []*Feature `json:"features"`
}

// Feature is a model for a feature.
type Feature struct {
	ID          string `json:"id"`
	Quota       map[string]int `json:"quota"`
}