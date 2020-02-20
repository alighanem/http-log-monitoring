package main

// IsInformational checks if status code is 1xx
func IsInformational(code int) bool {
	return code/100 == 1
}

// IsSuccess checks if status code is 2xx
func IsSuccess(code int) bool {
	return code/100 == 2
}

// IsRedirection checks if status code is 3xx
func IsRedirection(code int) bool {
	return code/100 == 3
}

// IsClientError checks if status code is 4xx:
func IsClientError(code int) bool {
	return code/100 == 4
}

// IsServerError checks if status code is 5xx
func IsServerError(code int) bool {
	return code/100 == 5
}
