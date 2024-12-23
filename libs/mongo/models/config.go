package models

type Config struct {
	MongoURL     string
	DatabaseName string
	Timeout      int
}

func NewConfig(mongoURL, databaseName string, timeout int) *Config {
	return &Config{
		MongoURL:     mongoURL,
		DatabaseName: databaseName,
		Timeout:      timeout,
	}
}
