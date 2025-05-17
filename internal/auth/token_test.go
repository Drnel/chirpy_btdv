package auth

import (
	"net/http"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer 99-77-ss-999")
	text, err := GetBearerToken(headers)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if text != "99-77-ss-999" {
		t.Errorf("Got wrong token: %s", text)
	}
}
