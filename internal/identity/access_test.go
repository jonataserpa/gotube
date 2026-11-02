package identity

import (
	"encoding/base64"
	"testing"
)

func TestParseAccessTokenRejectsUnsignedToken(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"1","email":"a@b.co","iss":"gotube","aud":"gotube-web"}`))
	if _, err := ParseAccessToken("test-secret-at-least-32-bytes-long", header+"."+payload+"."); err == nil {
		t.Fatal("accepted an unsigned token")
	}
}
