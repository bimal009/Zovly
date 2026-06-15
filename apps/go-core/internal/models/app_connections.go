package models

import (
	"time"
)

type AppConnections struct {
	ID         string `db:"id" json:"id"`
	BusinessID string `db:"business_id" json:"business_id"`

	Instagram bool `db:"instagram" json:"instagram"`
	Facebook  bool `db:"facebook" json:"facebook"`
	TikTok    bool `db:"tiktok" json:"tiktok"`
	WhatsApp  bool `db:"whatsapp" json:"whatsapp"`

	GoogleWorkspace bool `db:"google_workspace" json:"google_workspace"`
	StripeConnect   bool `db:"stripe_connect" json:"stripe_connect"`

	Fonepay bool `db:"fonepay" json:"fonepay"`
	Khalti  bool `db:"khalti" json:"khalti"`
	Esewa   bool `db:"esewa" json:"esewa"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
type ConnectionType string

const (
	ConnectionInstagram ConnectionType = "instagram"
	ConnectionFacebook  ConnectionType = "facebook"
	ConnectionTikTok    ConnectionType = "tiktok"
	ConnectionWhatsApp  ConnectionType = "whatsapp"

	ConnectionGoogleWorkspace ConnectionType = "google_workspace"
	ConnectionStripeConnect   ConnectionType = "stripe_connect"

	ConnectionFonepay ConnectionType = "fonepay"
	ConnectionKhalti  ConnectionType = "khalti"
	ConnectionEsewa   ConnectionType = "esewa"
)

var ConnectionColumns = map[ConnectionType]string{
	ConnectionInstagram:       "instagram",
	ConnectionFacebook:        "facebook",
	ConnectionTikTok:          "tiktok",
	ConnectionWhatsApp:        "whatsapp",
	ConnectionGoogleWorkspace: "google_workspace",
	ConnectionStripeConnect:   "stripe_connect",
	ConnectionFonepay:         "fonepay",
	ConnectionKhalti:          "khalti",
	ConnectionEsewa:           "esewa",
}

type FbPage struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
}

type FbPagesResponse struct {
	Data []FbPage `json:"data"`
}
type FbTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
