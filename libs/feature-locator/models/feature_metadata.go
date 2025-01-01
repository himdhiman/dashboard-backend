package models

type FeatureMetaData struct {
	FeatureName string `json:"feature_name"`
	Code 	  string `json:"code"`
	CreatedAt string `json:"created_at"`
	ModifiedAt string `json:"modified_at"`
	EffectiveDate string `json:"effective_date"`
	RemovalDate string `json:"removal_date"`
	Feature Feature `json:"feature"`
	Filters []FeatureFilter `json:"filters"`	
}

