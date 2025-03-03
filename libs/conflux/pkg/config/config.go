package config

type ApiClientConfig struct {
	ApiCodes []string
}

func NewApiClientConfig() *ApiClientConfig {
	return &ApiClientConfig{
		ApiCodes: []string{},
	}
}

func (c *ApiClientConfig) AddApiCode(code string) {
	c.ApiCodes = append(c.ApiCodes, code)
}

func (c *ApiClientConfig) GetApiCode(code string) string {
	for _, apiCode := range c.ApiCodes {
		if apiCode == code {
			return apiCode
		}
	}
	return ""
}

func (c *ApiClientConfig) GetAllApiCodes() []string {
	return c.ApiCodes
}
