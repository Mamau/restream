package middleware

import (
	"fmt"
	"net/http"
)

const addr = "http://0.0.0.0:8081"

func IsNginxRestreamRunning(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(addr)
		if err != nil {
			message := fmt.Sprintf("Nginx is not running, error: %v\n", err)
			http.Error(w, message, http.StatusBadRequest)
			return
		}

		isOk := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
		if !isOk {
			message := fmt.Sprintf("Nginx restream %v is not available %v\n", addr, resp.StatusCode)
			http.Error(w, message, http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
