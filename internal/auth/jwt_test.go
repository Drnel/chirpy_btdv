package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJwt(t *testing.T) {
	uid := uuid.New()
	signed_string, err := MakeJWT(uid, "Secret token string", time.Hour)
	if err != nil {
		t.Errorf("Couldnt make jwt: %s", err)
	}
	uid_2, err := ValidateJWT(signed_string, "Secret token string")
	if err != nil {
		t.Errorf("Couldnt validate jwt: %s", err)
	}
	if uid != uid_2 {
		t.Errorf("got Differednt uuids %s \n %s\n", uid, uid_2)
	}
}
