package constants

const (
	UNICOM_API_CODE = "UNICOM_SALTY"

	// API Codes
	API_CODE_UNICOM_FETCH_PRODUCTS = "FETCH_PRODUCTS"
)

const (
	BASE_URL         = ":baseURL"
	AUTH_TYPE        = ":Auth:Type"
	AUTH_PATH        = ":Auth:Path"
	AUTH_CREDENTIALS = ":Auth:Credentials"

	API_PATH       = ":Path"
	API_METHOD     = ":Method"
	API_RATE_LIMIT = ":RateLimit"
	API_TIMEOUT    = ":Timeout"
)

func GetBaseURLKey(apiCode string) string {
	return apiCode + BASE_URL
}

func GetAuthTypeKey(apiCode string) string {
	return apiCode + AUTH_TYPE
}

func GetAuthPathKey(apiCode string) string {
	return apiCode + AUTH_PATH
}

func GetAuthCredentialsKey(apiCode string) string {
	return apiCode + AUTH_CREDENTIALS
}

func GetApiPathKey(apiCode, endpointCode string) string {
	return apiCode + ":" + endpointCode + API_PATH
}

func GetApiMethodKey(apiCode, endpointCode string) string {
	return apiCode + ":" + endpointCode + API_METHOD
}

func GetApiRateLimitKey(apiCode, endpointCode string) string {
	return apiCode + ":" + endpointCode + API_RATE_LIMIT
}

func GetApiTimeoutKey(apiCode, endpointCode string) string {
	return apiCode + ":" + endpointCode + API_TIMEOUT
}
