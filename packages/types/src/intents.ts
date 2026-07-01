// intents.ts

export const Intent = {
  PRODUCT_SEARCH: "product_search",
  SERVICE_INQUIRY: "service_inquiry",
  EVENT_INQUIRY: "event_inquiry",
  ORDER_STATUS: "order_status",
  POLICY_FAQ: "policy_faq",
  BOOKING_CONFIRM: "booking_confirm",
  CART_ACTION: "cart_action",       // "add the black one", "remove it", "checkout"
  SMALL_TALK: "small_talk",
  UNKNOWN: "unknown",               // low confidence, fall back to clarifying question
} as const;

export type IntentType = typeof Intent[keyof typeof Intent];

// Entities extracted alongside intent — all optional, only present if mentioned
export interface ExtractedEntities {
  item?: string;          // "kurtha", "saree"
  color?: string;         // "black"
  size?: string;          // "M", "L"
  budget_min?: number;
  budget_max?: number;
  quantity?: number;
  date?: string;          // for events/services, ISO format if resolvable
  order_id?: string;
  product_ref?: string;   // "the black one" — resolved against last_shown_products, not stored raw
}

export interface SubIntent {
  intent: IntentType;
  text: string;            // the relevant slice of the original message
  entities: ExtractedEntities;
  confidence: number;      // 0-1
}

export interface IntentExtractionResult {
  sub_intents: SubIntent[];
}