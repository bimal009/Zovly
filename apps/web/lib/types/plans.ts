export type Plan = {
  id: string;
  name: string;

  paddle_product_id?: string | null;
  paddle_price_id_monthly?: string | null;
  paddle_price_id_yearly?: string | null;

  monthly_price: number;
  yearly_price: number;
  max_members: number;
  max_social_accounts: number;
  max_ai_replies_month: number;
  max_posts_month: number;
  max_leads: number;
  max_products: number;
  max_bookings_month: number;

  has_video_upload: boolean;
  has_multi_platform_post: boolean;
  has_post_analytics: boolean;
  has_ai_dm_replies: boolean;
  has_ai_comment_replies: boolean;
  has_ai_lead_scoring: boolean;
  has_ai_ad_suggestions: boolean;
  has_voice_transcription: boolean;
  has_image_understanding: boolean;
  has_bookings: boolean;
  has_inventory: boolean;
  has_payments: boolean;
  has_google_workspace: boolean;
  has_meta_ads: boolean;
  has_tiktok_ads: boolean;
  has_priority_support: boolean;

  is_active: boolean;

  created_at: string;
  updated_at: string;
};
