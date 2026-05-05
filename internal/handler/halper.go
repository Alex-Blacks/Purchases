package handler

import (
	"fmt"
	"net/http"
	"strconv"
)

func parseIntQuery(r *http.Request, key string) (int, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return 0, fmt.Errorf("%s required", key)
	}

	keyInt, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%s must be int", key)
	}
	return keyInt, err
}
