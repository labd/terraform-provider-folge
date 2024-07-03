package internal

import "net/http"

func cleanHeaders(headers http.Header, keep ...string) http.Header {
	for key := range headers {
		if !contains(keep, key) {
			headers.Del(key)
		}
	}
	return headers
}

func contains(keep []string, key string) bool {
	for _, k := range keep {
		if k == key {
			return true
		}
	}
	return false
}
