export type ConnectedPage = {
  id: string;

  business_id: string;
  app_name: string;

  access_token: string | null;
  refresh_token: string | null;
  token_expires_at: string | null;

  scopes: string[] | null;

  public_key: string | null;
  secret_key: string | null;
  merchant_id: string | null;

  platform_account_id: string;
  platform_account_name: string | null;
  platform_account_image: string | null;

  webhook_verify_token: string | null;
  webhook_subscribed_at: string | null;

  is_active: boolean;
  connected_at: string | null;
  disconnected_at: string | null;
  last_sync_at: string | null;

  error_message: string | null;

  created_at: string;
  updated_at: string;

  details?: PageDetails | null;
};

export type PageDetails = {
  id: string;
  name: string;
  about?: string;
  fan_count: number;
  followers_count: number;
  picture: { data: { url: string } };
  link: string;
  category: string;
};