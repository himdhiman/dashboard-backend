package constants

const (
	UNICOM_API_CODE = "UNICOM_SALTY"

	// API Codes
	API_CODE_UNICOM_FETCH_PRODUCTS    = "FETCH_PRODUCTS"
	API_CODE_UNICOM_CREATE_JOB        = "CREATE_EXPORT_JOB"
	API_CODE_UNICOM_EXPORT_JOB_STATUS = "EXPORT_JOB_STATUS"
	API_CODE_GET_INVENTORY_SNAPSHOT   = "GET_INVENTORY_SNAPSHOT"
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

const (
	EXPORT_JOB_CODE = "export_job_code"
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

func GetUnicomExportJobCode() string {
	return UNICOM_API_CODE + ":" + EXPORT_JOB_CODE
}
