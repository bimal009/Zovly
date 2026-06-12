package models

import "time"

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusDeclined InviteStatus = "declined"
	InviteStatusExpired  InviteStatus = "expired"
	InviteStatusRevoked  InviteStatus = "revoked"
)

type MemberInvite struct {
	ID           string       `db:"id"             json:"id"`
	BusinessID   string       `db:"business_id"    json:"business_id"`
	InvitedByID  *string      `db:"invited_by_id"  json:"invited_by_id,omitempty"`
	InvitedEmail string       `db:"invited_email"  json:"invited_email"`
	Role         MemberRole   `db:"role"           json:"role"`
	Token        string       `db:"token"          json:"-"`
	Status       InviteStatus `db:"status"         json:"status"`
	ExpiresAt    time.Time    `db:"expires_at"     json:"expires_at"`
	AcceptedAt   *time.Time   `db:"accepted_at"    json:"accepted_at,omitempty"`
	DeclinedAt   *time.Time   `db:"declined_at"    json:"declined_at,omitempty"`
	RevokedAt    *time.Time   `db:"revoked_at"     json:"revoked_at,omitempty"`
	RevokedByID  *string      `db:"revoked_by_id"  json:"revoked_by_id,omitempty"`
	CreatedAt    time.Time    `db:"created_at"     json:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at"     json:"updated_at"`
}
