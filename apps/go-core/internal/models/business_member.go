package models

import "time"

type MemberRole string

const (
	MemberRoleOwner   MemberRole = "owner"
	MemberRoleAdmin   MemberRole = "admin"
	MemberRoleManager MemberRole = "manager"
	MemberRoleStaff   MemberRole = "staff"
	MemberRoleViewer  MemberRole = "viewer"
)

type BusinessMember struct {
	ID         string     `db:"id" json:"id"`
	BusinessID string     `db:"business_id" json:"business_id"`
	UserID     string     `db:"user_id" json:"user_id"`
	Role       MemberRole `db:"role" json:"role"`

	CanManageContent bool `db:"can_manage_content" json:"can_manage_content"`
	CanViewAnalytics bool `db:"can_view_analytics" json:"can_view_analytics"`
	CanManageAds     bool `db:"can_manage_ads" json:"can_manage_ads"`

	CanReadDMs       bool `db:"can_read_dms" json:"can_read_dms"`
	CanReplyDMs      bool `db:"can_reply_dms" json:"can_reply_dms"`
	CanReadComments  bool `db:"can_read_comments" json:"can_read_comments"`
	CanReplyComments bool `db:"can_reply_comments" json:"can_reply_comments"`

	// Leads
	CanViewLeads   bool `db:"can_view_leads" json:"can_view_leads"`
	CanManageLeads bool `db:"can_manage_leads" json:"can_manage_leads"`

	// Bookings
	CanViewBookings   bool `db:"can_view_bookings" json:"can_view_bookings"`
	CanManageBookings bool `db:"can_manage_bookings" json:"can_manage_bookings"`

	// Inventory & Orders
	CanViewInventory   bool `db:"can_view_inventory" json:"can_view_inventory"`
	CanManageInventory bool `db:"can_manage_inventory" json:"can_manage_inventory"`
	CanViewOrders      bool `db:"can_view_orders" json:"can_view_orders"`

	// Settings
	CanManageSettings bool `db:"can_manage_settings" json:"can_manage_settings"`
	CanManageMembers  bool `db:"can_manage_members" json:"can_manage_members"`
	CanManageBilling  bool `db:"can_manage_billing" json:"can_manage_billing"`

	JoinedAt   time.Time  `db:"joined_at" json:"joined_at"`
	LastSeenAt *time.Time `db:"last_seen_at" json:"last_seen_at"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
