export interface FacebookWebhookPayload {
  object: "page" | "instagram" | "whatsapp_business_account" | string;
  entry: FacebookWebhookEntry[];
}

export interface FacebookWebhookEntry {
  id: string; // Page ID
  time: number; // unix timestamp
  changes?: FacebookWebhookChange[]; // "feed"/"mention"/"leadgen"/etc events
  messaging?: FacebookMessagingEvent[]; // Messenger events (not under `changes`)
}


export type FacebookWebhookChange =
  | { field: "feed"; value: FacebookFeedValue }
  | { field: "mention"; value: FacebookMentionValue }
  | { field: "leadgen"; value: FacebookLeadgenValue }
  | { field: "ratings"; value: FacebookRatingValue }
  | { field: string; value: Record<string, unknown> }; // fallback for any other field

export interface FacebookFeedValue {
  item: "post" | "comment" | "photo" | "video" | "status" | "like" | "share" | "reaction";
  verb: "add" | "edited" | "remove" | "edit";
  post_id?: string;
  comment_id?: string;
  parent_id?: string;
  sender_id?: string;
  sender_name?: string;
  message?: string;
  photo_id?: string;
  video_id?: string;
  created_time?: number;
  is_hidden?: boolean;
}

export interface FacebookMentionValue {
  item: "post" | "comment";
  post_id?: string;
  comment_id?: string;
  sender_id?: string;
  sender_name?: string;
  message?: string;
  created_time?: number;
}

export interface FacebookLeadgenValue {
  leadgen_id: string;
  page_id: string;
  form_id: string;
  adgroup_id?: string;
  ad_id?: string;
  created_time: number;
}

export interface FacebookRatingValue {
  item: "rating";
  verb: "add" | "edited" | "remove";
  review_id?: string;
  rating?: number;
  recommendation_type?: "positive" | "negative" | "no_recommendation";
  reviewer_id?: string;
  reviewer_name?: string;
  open_graph_story?: { message?: string };
  created_time?: number;
}



interface FacebookMessagingBase {
  sender: { id: string };
  recipient: { id: string };
  timestamp: number;
}

export interface FacebookMessageAttachment {
  type: "image" | "video" | "audio" | "file" | "template" | "fallback" | string;
  payload: {
    url?: string;
    sticker_id?: number;
    [key: string]: unknown;
  };
}

export interface FacebookMessageEvent extends FacebookMessagingBase {
  message: {
    mid: string;
    text?: string;
    is_echo?: boolean;
    app_id?: number; 
    attachments?: FacebookMessageAttachment[];
    quick_reply?: { payload: string };
    reply_to?: { mid: string };
  };
}

export interface FacebookMessageDeliveryEvent extends FacebookMessagingBase {
  delivery: {
    mids?: string[];
    watermark: number; 
  };
}

export interface FacebookMessageReadEvent extends FacebookMessagingBase {
  read: {
    watermark: number; 
  };
}

export interface FacebookMessagePostbackEvent extends FacebookMessagingBase {
  postback: {
    title: string;
    payload: string;
    referral?: FacebookReferral;
  };
}

export interface FacebookMessageReactionEvent extends FacebookMessagingBase {
  reaction: {
    mid: string;
    action: "react" | "unreact";
    reaction?: "smile" | "angry" | "sad" | "wow" | "love" | "like" | "dislike" | "other";
    emoji?: string;
  };
}

export interface FacebookReferral {
  ref?: string;
  source: "SHORTLINK" | "ADS" | "MESSENGER_CODE" | "DISCOVER_TAB" | string;
  type: "OPEN_THREAD";
  ads_context_data?: { ad_title?: string; photo_url?: string };
}

export interface FacebookMessageReferralEvent extends FacebookMessagingBase {
  referral: FacebookReferral;
}

export interface FacebookMessageOptinEvent extends FacebookMessagingBase {
  optin: {
    ref?: string;
    user_ref?: string;
  };
}

export interface FacebookHandoverEvent extends FacebookMessagingBase {
  pass_thread_control?: { new_owner_app_id: string; metadata?: string };
  take_thread_control?: { previous_owner_app_id: string; metadata?: string };
  request_thread_control?: { requested_owner_app_id: string; metadata?: string };
}

export type FacebookMessagingEvent =
  | FacebookMessageEvent
  | FacebookMessageDeliveryEvent
  | FacebookMessageReadEvent
  | FacebookMessagePostbackEvent
  | FacebookMessageReactionEvent
  | FacebookMessageReferralEvent
  | FacebookMessageOptinEvent
  | FacebookHandoverEvent;

export const isMessageEvent = (e: FacebookMessagingEvent): e is FacebookMessageEvent =>
  "message" in e;
export const isDeliveryEvent = (e: FacebookMessagingEvent): e is FacebookMessageDeliveryEvent =>
  "delivery" in e;
export const isReadEvent = (e: FacebookMessagingEvent): e is FacebookMessageReadEvent =>
  "read" in e;
export const isPostbackEvent = (e: FacebookMessagingEvent): e is FacebookMessagePostbackEvent =>
  "postback" in e;
export const isReactionEvent = (e: FacebookMessagingEvent): e is FacebookMessageReactionEvent =>
  "reaction" in e;
export const isReferralEvent = (e: FacebookMessagingEvent): e is FacebookMessageReferralEvent =>
  "referral" in e;
export const isOptinEvent = (e: FacebookMessagingEvent): e is FacebookMessageOptinEvent =>
  "optin" in e;
export const isHandoverEvent = (e: FacebookMessagingEvent): e is FacebookHandoverEvent =>
  "pass_thread_control" in e || "take_thread_control" in e || "request_thread_control" in e;