export type Plan = {
  id: string;
  name: string;

  paddleProductId: string | null;
  paddlePriceIdMonthly: string | null;
  paddlePriceIdYearly: string | null;

  monthlyPrice: number;
  yearlyPrice: number;

  maxMembers: number;
  maxSocialAccounts: number;
  maxAiRepliesMonth: number;
  maxPostsMonth: number;
  maxLeads: number;
  maxProducts: number;
  maxBookingsMonth: number;

  hasVideoUpload: boolean;
  hasMultiPlatformPost: boolean;
  hasPostAnalytics: boolean;
  hasAiDmReplies: boolean;
  hasAiCommentReplies: boolean;
  hasAiLeadScoring: boolean;
  hasAiAdSuggestions: boolean;
  hasVoiceTranscription: boolean;
  hasImageUnderstanding: boolean;
  hasBookings: boolean;
  hasInventory: boolean;
  hasPayments: boolean;
  hasGoogleWorkspace: boolean;
  hasMetaAds: boolean;
  hasTikTokAds: boolean;
  hasPrioritySupport: boolean;

  aiReplyOveragePriceUsdPer500: number | null;

  isActive: boolean;

  createdAt: Date;
  updatedAt: Date;
};
