export type Platform = "facebook" | "instagram";

export type Conversation = {
  id: string;
  business_id: string;
  platform: Platform;
  thread_id: string;
  contact_id: string;
  contact_name: string | null;
  contact_username: string | null;
  contact_avatar_url: string | null;
  last_message_at: string | null;
  created_at: string;
};

export type MessageDirection = "in" | "out";
export type MessageSender = "ai" | "human";
export type MessageStatus = "pending" | "sent" | "failed" | "skipped";
export type MessageMediaType =
  | "image"
  | "video"
  | "audio"
  | "document"
  | "link";

export type Message = {
  id: string;
  conversation_id: string;
  business_id: string;
  direction: MessageDirection;
  sent_by: MessageSender | null;
  content: string | null;
  media_url: string | null;
  media_type: MessageMediaType | null;
  is_vectorized: boolean;
  status: MessageStatus | null;
  error_message: string | null;
  sent_to_platform_at: string | null;
  platform_message_id: string | null;
  sent_at: string;
};
