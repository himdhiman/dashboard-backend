package models

type FilterDefination struct {
	Name             string `json:"name"`
	PropertyPath     string `json:"property_path"`
	IsInclusionFiter bool   `json:"is_inclusion_filter"`
}
