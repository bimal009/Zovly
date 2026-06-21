package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// verifyMetaSignature validates the X-Hub-Signature-256 header against the raw
// request body using the given app secret. Shared by the Facebook and Instagram
// webhook handlers — each passes its own app secret (META_APP_SECRET vs
// IG_APP_SECRET), since Meta signs each product's webhook with its own secret.
func verifyMetaSignature(body []byte, header, appSecret string) bool {
	const prefix = "sha256="

	if !strings.HasPrefix(header, prefix) {
		return false
	}

	got, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)

	return hmac.Equal(mac.Sum(nil), got)
}
