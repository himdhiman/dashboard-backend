package models

type FeatureFilter struct {
	Type     string   `json:"type"`
	Priority int      `json:"priority"`
	Tags     []string `json:"tags"`
	Apply    bool     `json:"apply"`
}
