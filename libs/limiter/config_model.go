package limiter

import "time"

type AlgorithmType string

const (
	FixedWindow   AlgorithmType = "fixed_window"
	SlidingWindow AlgorithmType = "sliding_window"
	TokenBucket   AlgorithmType = "token_bucket"
	LeakyBucket   AlgorithmType = "leaky_bucket"
)

type DefaultConfig struct {
	Algorithm   AlgorithmType `json:"algorithm" bson:"algorithm"`
	MaxRequests int           `json:"max_requests" bson:"max_requests"`
	TimeWindow  time.Duration `json:"time_window" bson:"time_window"`
}

type EndpointConfig struct {
	Endpoint  string        `json:"endpoint" bson:"endpoint"`
	Algorithm AlgorithmType `json:"algorithm" bson:"algorithm"`
}

type RateLimitConfig struct {
	ServiceName     string           `json:"service_name" bson:"service_name"`
	DefaultConfig   []DefaultConfig  `json:"default_config" bson:"default_config"`
	EndpointConfigs []EndpointConfig `json:"endpoint_configs" bson:"endpoint_configs"`
}
