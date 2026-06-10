package models

type ImageAuthTokenResponse struct {
	Signature string `json:"signature"`
	Expire    int64  `json:"expire"`
	Token     string `json:"token"`
}
