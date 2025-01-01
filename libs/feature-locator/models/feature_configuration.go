package models

type FeatureConfiguration struct {
	Name                  string             `json:"name"`
	RequiresAtleast1Value bool               `json:"requires_atleast_1_value"`
	ReturnSingleValue     bool               `json:"return_single_value"`
	FeatureMetaData       FeatureMetaData    `json:"feature_metadata"`
	FilterDefination      []FilterDefination `json:"filter_defination"`
}
