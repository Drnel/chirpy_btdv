package auth

import (
	"errors"
	"net/http"
)

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if token == "" {
		return "", errors.New("No Auth token received")
	}
	if len(token) < 7 {
		return "", errors.New("Bad auth token")
	}
	return string([]byte(token)[7:]), nil
}
