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

export type ConnectedPage = {
  id: string;
  business_id: string;
  app_name: string;
  platform_account_id: string | null;
  platform_account_name: string | null;
  is_active: boolean;
  connected_at: string | null;
  disconnected_at: string | null;
  token_expires_at: string | null;
  scopes: string[];
  webhook_subscribed_at: string | null;
  last_sync_at: string | null;
  error_message: string | null;
  created_at: string;
  updated_at: string;
  details?: PageDetails | null;
};

export type FacebookConnectionStatus = {
  connected: boolean;
  pages: ConnectedPage[];
};

export type InstagramConnectionStatus = {
  connected: boolean;
  account: ConnectedPage | null;
};

export type BusinessAppConnections = {
  id: string;
  business_id: string;
  instagram: boolean;
  facebook: boolean;
  tiktok: boolean;
  whatsapp: boolean;
  google_workspace: boolean;
  stripe_connect: boolean;
  fonepay: boolean;
  khalti: boolean;
  esewa: boolean;
  created_at: string;
  updated_at: string;
};
