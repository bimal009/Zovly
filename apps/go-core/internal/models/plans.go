package models

import "time"

type Plan struct {
	ID   string `db:"id"   json:"id"`
	Name string `db:"name" json:"name"`

	// ── Paddle IDs ────────────────────────────────────────────────────────────
	PaddleProductID      *string `db:"paddle_product_id"       json:"paddle_product_id,omitempty"`
	PaddlePriceIDMonthly *string `db:"paddle_price_id_monthly" json:"paddle_price_id_monthly,omitempty"`
	PaddlePriceIDYearly  *string `db:"paddle_price_id_yearly"  json:"paddle_price_id_yearly,omitempty"`

	// ── Pricing (cents USD) ───────────────────────────────────────────────────
	MonthlyPrice int `db:"monthly_price" json:"monthly_price"`
	YearlyPrice  int `db:"yearly_price"  json:"yearly_price"`

	// ── Limits (-1 = unlimited) ───────────────────────────────────────────────
	MaxMembers        int `db:"max_members"          json:"max_members"`
	MaxSocialAccounts int `db:"max_social_accounts"  json:"max_social_accounts"`
	MaxAiRepliesMonth int `db:"max_ai_replies_month" json:"max_ai_replies_month"`
	MaxPostsMonth     int `db:"max_posts_month"      json:"max_posts_month"`
	MaxLeads          int `db:"max_leads"            json:"max_leads"`
	MaxProducts       int `db:"max_products"         json:"max_products"`
	MaxBookingsMonth  int `db:"max_bookings_month"   json:"max_bookings_month"`

	// ── Feature flags ─────────────────────────────────────────────────────────
	HasVideoUpload        bool `db:"has_video_upload"         json:"has_video_upload"`
	HasMultiPlatformPost  bool `db:"has_multi_platform_post"  json:"has_multi_platform_post"`
	HasPostAnalytics      bool `db:"has_post_analytics"       json:"has_post_analytics"`
	HasAiDmReplies        bool `db:"has_ai_dm_replies"        json:"has_ai_dm_replies"`
	HasAiCommentReplies   bool `db:"has_ai_comment_replies"   json:"has_ai_comment_replies"`
	HasAiLeadScoring      bool `db:"has_ai_lead_scoring"      json:"has_ai_lead_scoring"`
	HasAiAdSuggestions    bool `db:"has_ai_ad_suggestions"    json:"has_ai_ad_suggestions"`
	HasVoiceTranscription bool `db:"has_voice_transcription"  json:"has_voice_transcription"`
	HasImageUnderstanding bool `db:"has_image_understanding"  json:"has_image_understanding"`
	HasBookings           bool `db:"has_bookings"             json:"has_bookings"`
	HasInventory          bool `db:"has_inventory"            json:"has_inventory"`
	HasNepalPayments      bool `db:"has_nepal_payments"       json:"has_nepal_payments"`
	HasGoogleWorkspace    bool `db:"has_google_workspace"     json:"has_google_workspace"`
	HasMetaAds            bool `db:"has_meta_ads"             json:"has_meta_ads"`
	HasTikTokAds          bool `db:"has_tiktok_ads"           json:"has_tiktok_ads"`
	HasWhiteLabel         bool `db:"has_white_label"          json:"has_white_label"`
	HasApiAccess          bool `db:"has_api_access"           json:"has_api_access"`
	HasPrioritySupport    bool `db:"has_priority_support"     json:"has_priority_support"`

	// ── Overage (nil = hard block) ────────────────────────────────────────────
	AiReplyOveragePriceUsdPer500 *int `db:"ai_reply_overage_price_usd_per_500" json:"ai_reply_overage_price_usd_per_500,omitempty"`

	IsActive bool `db:"is_active" json:"is_active"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// IsUnlimited returns true if the given limit value means "no limit"
func IsUnlimited(limit int) bool {
	return limit == -1
}

// CanUse checks a usage count against a plan limit, respecting -1 = unlimited
func (p *Plan) CanUse(limit int, current int) bool {
	if IsUnlimited(limit) {
		return true
	}
	return current < limit
}

// ── Feature check helpers (used in middleware) ────────────────────────────────

func (p *Plan) CanPost() bool              { return p.HasVideoUpload || p.HasMultiPlatformPost }
func (p *Plan) CanUseAI() bool             { return p.HasAiDmReplies }
func (p *Plan) CanUseBookings() bool       { return p.HasBookings }
func (p *Plan) CanUseInventory() bool      { return p.HasInventory }
func (p *Plan) CanUseGoogleWorkspace() bool { return p.HasGoogleWorkspace }
func (p *Plan) CanRunAds() bool            { return p.HasMetaAds || p.HasTikTokAds }

// ── Pricing helpers ───────────────────────────────────────────────────────────

// MonthlyPriceDollars returns the monthly price as a float (e.g. 1900 → 19.00)
func (p *Plan) MonthlyPriceDollars() float64 {
	return float64(p.MonthlyPrice) / 100
}

// YearlyPriceDollars returns the yearly price as a float
func (p *Plan) YearlyPriceDollars() float64 {
	return float64(p.YearlyPrice) / 100
}

// YearlySavingsDollars returns how much you save vs paying monthly for 12 months
func (p *Plan) YearlySavingsDollars() float64 {
	return float64(p.MonthlyPrice*12-p.YearlyPrice) / 100
}

// ── Partial update structs ────────────────────────────────────────────────────

type PlanPaddleUpdate struct {
	PaddleProductID      *string
	PaddlePriceIDMonthly *string
	PaddlePriceIDYearly  *string
}

type PlanPricingUpdate struct {
	MonthlyPrice int
	YearlyPrice  int
}