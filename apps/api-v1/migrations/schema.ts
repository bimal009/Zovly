import { pgEnum, pgTable, text, uuid, timestamp, boolean, integer, vector, jsonb, numeric, customType, index, foreignKey, primaryKey, unique } from "drizzle-orm/pg-core"
import { sql } from "drizzle-orm"

export const userRole = pgEnum("user_role", ["vendor", "admin", "user"])
export const businessType = pgEnum("business_type", ["product", "service", "both"])
export const plan = pgEnum("plan", ["starter", "growth", "pro", "agency"])
export const inviteStatus = pgEnum("invite_status", ["pending", "accepted", "declined", "expired", "revoked"])
export const memberRole = pgEnum("member_role", ["owner", "admin", "manager", "staff", "viewer"])
export const permissionAction = pgEnum("permission_action", ["read", "write", "delete", "manage"])
export const billingCycle = pgEnum("billing_cycle", ["monthly", "yearly"])
export const paymentMethod = pgEnum("payment_method", ["esewa", "khalti", "fonepay", "bank_transfer", "cash", "other"])
export const planStatus = pgEnum("plan_status", ["active", "trialing", "past_due", "paused", "cancelled", "expired"])
export const paymentStatus = pgEnum("payment_status", ["paid", "refunded", "partially_refunded", "failed"])
export const productStatus = pgEnum("product_status", ["active", "inactive", "archived"])
export const billingInterval = pgEnum("billing_interval", ["weekly", "monthly", "quarterly", "yearly"])
export const serviceStatus = pgEnum("service_status", ["active", "inactive", "archived"])
export const serviceType = pgEnum("service_type", ["appointment", "membership", "class", "package"])
export const messageDirection = pgEnum("message_direction", ["in", "out"])
export const messageMediaType = pgEnum("message_media_type", ["image", "video", "audio", "document", "link"])
export const messageSender = pgEnum("message_sender", ["ai", "human"])
export const platform = pgEnum("platform", ["instagram", "facebook", "whatsapp", "tiktok"])
export const knowledgeSourceType = pgEnum("knowledge_source_type", ["faq", "policy", "post", "product", "service"])
export const messageStatus = pgEnum("message_status", ["pending", "sent", "failed", "skipped"])


export const account = pgTable("account", {
	id: text().primaryKey(),
	accountId: text("account_id").notNull(),
	providerId: text("provider_id").notNull(),
	userId: text("user_id").notNull().references(() => user.id, { onDelete: "cascade" } ),
	accessToken: text("access_token"),
	refreshToken: text("refresh_token"),
	idToken: text("id_token"),
	accessTokenExpiresAt: timestamp("access_token_expires_at"),
	refreshTokenExpiresAt: timestamp("refresh_token_expires_at"),
	scope: text(),
	password: text(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").notNull(),
}, (table) => [
	index("account_accountId_idx").using("btree", table.accountId.asc().nullsLast()),
	index("account_provider_idx").using("btree", table.providerId.asc().nullsLast()),
	index("account_user_idx").using("btree", table.userId.asc().nullsLast()),
]);

export const appConnections = pgTable("app_connections", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	instagram: boolean().default(false).notNull(),
	facebook: boolean().default(false).notNull(),
	tiktok: boolean().default(false).notNull(),
	whatsapp: boolean().default(false).notNull(),
	googleWorkspace: boolean("google_workspace").default(false).notNull(),
	stripeConnect: boolean("stripe_connect").default(false).notNull(),
	fonepay: boolean().default(false).notNull(),
	khalti: boolean().default(false).notNull(),
	esewa: boolean().default(false).notNull(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("app_conn_business_idx").using("btree", table.businessId.asc().nullsLast()),
	unique("app_connections_business_id_unique").on(table.businessId),]);

export const appCredentials = pgTable("app_credentials", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	appName: text("app_name").notNull(),
	accessToken: text("access_token"),
	refreshToken: text("refresh_token"),
	tokenExpiresAt: timestamp("token_expires_at"),
	scopes: text().array(),
	publicKey: text("public_key"),
	secretKey: text("secret_key"),
	merchantId: text("merchant_id"),
	platformAccountId: text("platform_account_id"),
	platformAccountName: text("platform_account_name"),
	webhookVerifyToken: text("webhook_verify_token"),
	webhookSubscribedAt: timestamp("webhook_subscribed_at"),
	isActive: boolean("is_active").default(true).notNull(),
	connectedAt: timestamp("connected_at"),
	disconnectedAt: timestamp("disconnected_at"),
	lastSyncAt: timestamp("last_sync_at"),
	errorMessage: text("error_message"),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("app_cred_app_name_idx").using("btree", table.appName.asc().nullsLast()),
	index("app_cred_business_idx").using("btree", table.businessId.asc().nullsLast()),
	unique("app_cred_app_uq").on(table.businessId, table.appName),]);

export const business = pgTable("business", {
	id: uuid().defaultRandom().primaryKey(),
	name: text().notNull(),
	description: text(),
	logo: text(),
	website: text(),
	phone: text(),
	address: text(),
	city: text(),
	country: text().default("NPL"),
	type: businessType().default("service").notNull(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("business_type_idx").using("btree", table.type.asc().nullsLast()),
]);

export const businessMembers = pgTable("business_members", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	userId: text("user_id").notNull().references(() => user.id, { onDelete: "cascade" } ),
	role: memberRole().default("viewer").notNull(),
	canManageContent: boolean("can_manage_content").default(false).notNull(),
	canViewAnalytics: boolean("can_view_analytics").default(false).notNull(),
	canManageAds: boolean("can_manage_ads").default(false).notNull(),
	canReadDms: boolean("can_read_dms").default(false).notNull(),
	canReplyDms: boolean("can_reply_dms").default(false).notNull(),
	canReadComments: boolean("can_read_comments").default(false).notNull(),
	canReplyComments: boolean("can_reply_comments").default(false).notNull(),
	canViewLeads: boolean("can_view_leads").default(false).notNull(),
	canManageLeads: boolean("can_manage_leads").default(false).notNull(),
	canViewBookings: boolean("can_view_bookings").default(false).notNull(),
	canManageBookings: boolean("can_manage_bookings").default(false).notNull(),
	canViewInventory: boolean("can_view_inventory").default(false).notNull(),
	canManageInventory: boolean("can_manage_inventory").default(false).notNull(),
	canViewOrders: boolean("can_view_orders").default(false).notNull(),
	canManageSettings: boolean("can_manage_settings").default(false).notNull(),
	canManageMembers: boolean("can_manage_members").default(false).notNull(),
	canManageBilling: boolean("can_manage_billing").default(false).notNull(),
	joinedAt: timestamp("joined_at").default(sql`now()`).notNull(),
	lastSeenAt: timestamp("last_seen_at"),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("biz_member_business_idx").using("btree", table.businessId.asc().nullsLast()),
	index("biz_member_role_idx").using("btree", table.role.asc().nullsLast()),
	index("biz_member_user_idx").using("btree", table.userId.asc().nullsLast()),
	unique("uq_business_member").on(table.businessId, table.userId),]);

export const businessSubscriptions = pgTable("business_subscriptions", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	planId: uuid("plan_id").notNull().references(() => plans.id),
	paddleSubscriptionId: text("paddle_subscription_id"),
	paddleCustomerId: text("paddle_customer_id"),
	paddlePriceId: text("paddle_price_id"),
	billingCycle: billingCycle("billing_cycle").default("monthly").notNull(),
	status: planStatus().default("trialing").notNull(),
	aiRepliesUsed: integer("ai_replies_used").default(0).notNull(),
	postsUsed: integer("posts_used").default(0).notNull(),
	usageResetAt: timestamp("usage_reset_at"),
	trialStartedAt: timestamp("trial_started_at"),
	trialEndsAt: timestamp("trial_ends_at"),
	currentPeriodStart: timestamp("current_period_start"),
	currentPeriodEnd: timestamp("current_period_end"),
	cancelAtPeriodEnd: boolean("cancel_at_period_end").default(false).notNull(),
	cancelledAt: timestamp("cancelled_at"),
	pausedAt: timestamp("paused_at"),
	notes: text(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("sub_business_idx").using("btree", table.businessId.asc().nullsLast()),
	index("sub_paddle_customer_idx").using("btree", table.paddleCustomerId.asc().nullsLast()),
	index("sub_paddle_sub_idx").using("btree", table.paddleSubscriptionId.asc().nullsLast()),
	index("sub_plan_idx").using("btree", table.planId.asc().nullsLast()),
	index("sub_status_idx").using("btree", table.status.asc().nullsLast()),
	unique("business_subscriptions_business_id_unique").on(table.businessId),	unique("business_subscriptions_paddle_subscription_id_unique").on(table.paddleSubscriptionId),]);

export const categories = pgTable("categories", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	name: text().notNull(),
	description: text(),
	slug: text(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("categories_business_idx").using("btree", table.businessId.asc().nullsLast()),
	unique("categories_business_name_uq").on(table.businessId, table.name),]);

export const conversations = pgTable("conversations", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	platform: platform().notNull(),
	threadId: text("thread_id").notNull(),
	contactId: text("contact_id").notNull(),
	contactName: text("contact_name"),
	contactUsername: text("contact_username"),
	contactAvatarUrl: text("contact_avatar_url"),
	lastMessageAt: timestamp("last_message_at").default(sql`now()`).notNull(),
	activeProductId: uuid("active_product_id").references(() => products.id, { onDelete: "set null" } ),
	activeProductAt: timestamp("active_product_at"),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
}, (table) => [
	index("conv_business_platform_idx").using("btree", table.businessId.asc().nullsLast(), table.platform.asc().nullsLast()),
	unique("conv_thread_uq").on(table.businessId, table.platform, table.threadId),]);

export const faqs = pgTable("faqs", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	question: text().notNull(),
	answer: text().notNull(),
	isActive: boolean("is_active").default(true).notNull(),
	sortOrder: integer("sort_order").default(0).notNull(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("faq_business_idx").using("btree", table.businessId.asc().nullsLast()),
]);

export const knowledgeChunks = pgTable("knowledge_chunks", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	sourceType: knowledgeSourceType("source_type").notNull(),
	sourceId: uuid("source_id").notNull(),
	chunkIndex: integer("chunk_index").default(0).notNull(),
	content: text().notNull(),
	embedding: vector({ dimensions: 1024 }).notNull(),
	metadata: jsonb(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
}, (table) => [
	index("kc_business_idx").using("btree", table.businessId.asc().nullsLast()),
	index("kc_hnsw_idx").using("hnsw", table.embedding.asc().nullsLast().op("vector_cosine_ops")),
	index("kc_source_idx").using("btree", table.businessId.asc().nullsLast(), table.sourceType.asc().nullsLast(), table.sourceId.asc().nullsLast()),
	unique("kc_source_chunk_uq").on(table.sourceType, table.sourceId, table.chunkIndex),]);

export const memberInvites = pgTable("member_invites", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	invitedById: text("invited_by_id").references(() => user.id, { onDelete: "set null" } ),
	invitedEmail: text("invited_email").notNull(),
	role: memberRole().default("viewer").notNull(),
	token: text().notNull(),
	status: inviteStatus().default("pending").notNull(),
	expiresAt: timestamp("expires_at").notNull(),
	acceptedAt: timestamp("accepted_at"),
	declinedAt: timestamp("declined_at"),
	revokedAt: timestamp("revoked_at"),
	revokedById: text("revoked_by_id").references(() => user.id, { onDelete: "set null" } ),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("invite_business_idx").using("btree", table.businessId.asc().nullsLast()),
	index("invite_email_idx").using("btree", table.invitedEmail.asc().nullsLast()),
	index("invite_status_idx").using("btree", table.status.asc().nullsLast()),
	index("invite_token_idx").using("btree", table.token.asc().nullsLast()),
	unique("member_invites_token_unique").on(table.token),]);

export const messageEmbeddings = pgTable("message_embeddings", {
	id: uuid().defaultRandom().primaryKey(),
	messageId: uuid("message_id").notNull().references(() => messages.id, { onDelete: "cascade" } ),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	conversationId: uuid("conversation_id").notNull().references(() => conversations.id, { onDelete: "cascade" } ),
	content: text().notNull(),
	embedding: vector({ dimensions: 1024 }).notNull(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
}, (table) => [
	index("msg_emb_business_idx").using("btree", table.businessId.asc().nullsLast()),
	index("msg_emb_conversation_idx").using("btree", table.conversationId.asc().nullsLast()),
	index("msg_emb_hnsw_idx").using("hnsw", table.embedding.asc().nullsLast().op("vector_cosine_ops")),
	unique("message_embeddings_message_id_unique").on(table.messageId),]);

export const messages = pgTable("messages", {
	id: uuid().defaultRandom().primaryKey(),
	conversationId: uuid("conversation_id").notNull().references(() => conversations.id, { onDelete: "cascade" } ),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	direction: messageDirection().notNull(),
	sentBy: messageSender("sent_by"),
	content: text(),
	mediaUrl: text("media_url"),
	mediaType: messageMediaType("media_type"),
	isVectorized: boolean("is_vectorized").default(false).notNull(),
	sentAt: timestamp("sent_at").default(sql`now()`).notNull(),
	status: messageStatus(),
	errorMessage: text("error_message"),
	sentToPlatformAt: timestamp("sent_to_platform_at"),
	platformMessageId: text("platform_message_id"),
}, (table) => [
	index("msg_conversation_idx").using("btree", table.conversationId.asc().nullsLast()),
	index("msg_vectorize_idx").using("btree", table.businessId.asc().nullsLast(), table.isVectorized.asc().nullsLast()),
]);

export const paymentRecords = pgTable("payment_records", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	subscriptionId: uuid("subscription_id").references(() => businessSubscriptions.id, { onDelete: "set null" } ),
	planId: uuid("plan_id").references(() => plans.id),
	billingCycle: billingCycle("billing_cycle").notNull(),
	paddleTransactionId: text("paddle_transaction_id").notNull(),
	paddleSubscriptionId: text("paddle_subscription_id"),
	paddleCustomerId: text("paddle_customer_id"),
	amount: integer().notNull(),
	currency: text().default("USD").notNull(),
	periodStart: timestamp("period_start").notNull(),
	periodEnd: timestamp("period_end").notNull(),
	status: paymentStatus().default("paid").notNull(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("payment_business_idx").using("btree", table.businessId.asc().nullsLast()),
	index("payment_paddle_sub_idx").using("btree", table.paddleSubscriptionId.asc().nullsLast()),
	index("payment_paddle_txn_idx").using("btree", table.paddleTransactionId.asc().nullsLast()),
	index("payment_status_idx").using("btree", table.status.asc().nullsLast()),
	index("payment_sub_idx").using("btree", table.subscriptionId.asc().nullsLast()),
	unique("payment_records_paddle_transaction_id_unique").on(table.paddleTransactionId),]);

export const plans = pgTable("plans", {
	id: uuid().defaultRandom().primaryKey(),
	name: text().notNull(),
	paddleProductId: text("paddle_product_id"),
	paddlePriceIdMonthly: text("paddle_price_id_monthly"),
	paddlePriceIdYearly: text("paddle_price_id_yearly"),
	monthlyPrice: integer("monthly_price").notNull(),
	yearlyPrice: integer("yearly_price").notNull(),
	maxMembers: integer("max_members").notNull(),
	maxSocialAccounts: integer("max_social_accounts").notNull(),
	maxAiRepliesMonth: integer("max_ai_replies_month").notNull(),
	maxPostsMonth: integer("max_posts_month").notNull(),
	maxLeads: integer("max_leads").notNull(),
	maxProducts: integer("max_products").notNull(),
	maxBookingsMonth: integer("max_bookings_month").notNull(),
	hasVideoUpload: boolean("has_video_upload").default(false).notNull(),
	hasMultiPlatformPost: boolean("has_multi_platform_post").default(false).notNull(),
	hasPostAnalytics: boolean("has_post_analytics").default(false).notNull(),
	hasAiDmReplies: boolean("has_ai_dm_replies").default(false).notNull(),
	hasAiCommentReplies: boolean("has_ai_comment_replies").default(false).notNull(),
	hasAiLeadScoring: boolean("has_ai_lead_scoring").default(false).notNull(),
	hasAiAdSuggestions: boolean("has_ai_ad_suggestions").default(false).notNull(),
	hasVoiceTranscription: boolean("has_voice_transcription").default(false).notNull(),
	hasImageUnderstanding: boolean("has_image_understanding").default(false).notNull(),
	hasBookings: boolean("has_bookings").default(false).notNull(),
	hasInventory: boolean("has_inventory").default(false).notNull(),
	hasPayments: boolean("has_payments").default(false).notNull(),
	hasGoogleWorkspace: boolean("has_google_workspace").default(false).notNull(),
	hasMetaAds: boolean("has_meta_ads").default(false).notNull(),
	hasTiktokAds: boolean("has_tiktok_ads").default(false).notNull(),
	hasPrioritySupport: boolean("has_priority_support").default(false).notNull(),
	aiReplyOveragePriceUsdPer500: integer("ai_reply_overage_price_usd_per_500"),
	isActive: boolean("is_active").default(true).notNull(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("plan_active_idx").using("btree", table.isActive.asc().nullsLast()),
	index("plan_name_idx").using("btree", table.name.asc().nullsLast()),
	index("plan_paddle_monthly_idx").using("btree", table.paddlePriceIdMonthly.asc().nullsLast()),
	index("plan_paddle_product_idx").using("btree", table.paddleProductId.asc().nullsLast()),
	index("plan_paddle_yearly_idx").using("btree", table.paddlePriceIdYearly.asc().nullsLast()),
	unique("plans_name_unique").on(table.name),	unique("plans_paddle_price_id_monthly_unique").on(table.paddlePriceIdMonthly),	unique("plans_paddle_price_id_yearly_unique").on(table.paddlePriceIdYearly),	unique("plans_paddle_product_id_unique").on(table.paddleProductId),]);

export const policies = pgTable("policies", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	title: text().notNull(),
	content: text().notNull(),
	isActive: boolean("is_active").default(true).notNull(),
	sortOrder: integer("sort_order").default(0).notNull(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("policy_business_idx").using("btree", table.businessId.asc().nullsLast()),
]);

export const productVariants = pgTable("product_variants", {
	id: uuid().defaultRandom().primaryKey(),
	productId: uuid("product_id").notNull().references(() => products.id, { onDelete: "cascade" } ),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	name: text().notNull(),
	sku: text(),
	attributes: jsonb(),
	price: numeric({ precision: 12, scale: 2 }),
	discount: integer(),
	stockQty: integer("stock_qty").default(0).notNull(),
	lowStockThreshold: integer("low_stock_threshold").default(5),
	images: text().array().default([]),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
	costPrice: numeric("cost_price", { precision: 12, scale: 2 }),
}, (table) => [
	index("pv_business_idx").using("btree", table.businessId.asc().nullsLast()),
	index("pv_product_idx").using("btree", table.productId.asc().nullsLast()),
	index("pv_sku_idx").using("btree", table.sku.asc().nullsLast()),
	unique("pv_product_name_uq").on(table.productId, table.name),]);

export const products = pgTable("products", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	name: text().notNull(),
	description: text(),
	sku: text(),
	status: productStatus().default("active").notNull(),
	price: numeric({ precision: 12, scale: 2 }).notNull(),
	costPrice: numeric("cost_price", { precision: 12, scale: 2 }),
	currency: text().default("NPR").notNull(),
	stockQty: integer("stock_qty").default(0).notNull(),
	lowStockThreshold: integer("low_stock_threshold").default(5),
	images: text().array().default([]),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
	discount: integer().default(0).notNull(),
	categoryId: uuid("category_id").references(() => categories.id, { onDelete: "set null" } ),
	tags: text().array().default([]),
	attributes: jsonb(),
	searchTsv: customType({ dataType: () => 'tsvector' })("search_tsv").generatedAlwaysAs(sql`(to_tsvector('simple'::regconfig, ((COALESCE(name, ''::text) || ' '::text) || COALESCE(description, ''::text))) || array_to_tsvector(array_remove(COALESCE(tags, '{}'::text[]), NULL::text)))`),
}, (table) => [
	index("idx_products_name_trgm").using("gin", table.name.asc().nullsLast().op("gin_trgm_ops")),
	index("idx_products_search_tsv").using("gin", table.searchTsv.asc().nullsLast()),
	index("products_business_id_idx").using("btree", table.businessId.asc().nullsLast()),
	index("products_business_status_idx").using("btree", table.businessId.asc().nullsLast(), table.status.asc().nullsLast()),
	index("products_category_id_idx").using("btree", table.categoryId.asc().nullsLast()),
	index("products_sku_idx").using("btree", table.sku.asc().nullsLast()),
	index("products_status_idx").using("btree", table.status.asc().nullsLast()),
]);

export const services = pgTable("services", {
	id: uuid().defaultRandom().primaryKey(),
	businessId: uuid("business_id").notNull().references(() => business.id, { onDelete: "cascade" } ),
	type: serviceType().default("appointment").notNull(),
	status: serviceStatus().default("active").notNull(),
	name: text().notNull(),
	description: text(),
	price: integer().notNull(),
	costPrice: integer("cost_price"),
	mrp: integer(),
	currency: text().default("NPR").notNull(),
	requiresDeposit: boolean("requires_deposit").default(false).notNull(),
	depositAmount: integer("deposit_amount"),
	durationMin: integer("duration_min"),
	bufferMin: integer("buffer_min").default(0),
	maxAdvanceDays: integer("max_advance_days").default(30),
	googleCalendarId: text("google_calendar_id"),
	maxConcurrent: integer("max_concurrent").default(1),
	billingInterval: billingInterval("billing_interval"),
	trialDays: integer("trial_days").default(0),
	sessionCount: integer("session_count"),
	validityDays: integer("validity_days"),
	images: text().array().default([]),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
	features: jsonb().default([]).notNull(),
}, (table) => [
	index("services_business_id_idx").using("btree", table.businessId.asc().nullsLast()),
	index("services_business_status_idx").using("btree", table.businessId.asc().nullsLast(), table.status.asc().nullsLast()),
	index("services_status_idx").using("btree", table.status.asc().nullsLast()),
	index("services_type_idx").using("btree", table.type.asc().nullsLast()),
]);

export const session = pgTable("session", {
	id: text().primaryKey(),
	expiresAt: timestamp("expires_at").notNull(),
	token: text().notNull(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").notNull(),
	ipAddress: text("ip_address"),
	userAgent: text("user_agent"),
	userId: text("user_id").notNull().references(() => user.id, { onDelete: "cascade" } ),
}, (table) => [
	index("session_userId_idx").using("btree", table.userId.asc().nullsLast()),
	unique("session_token_unique").on(table.token),]);

export const user = pgTable("user", {
	id: text().primaryKey(),
	name: text().notNull(),
	email: text().notNull(),
	emailVerified: boolean("email_verified").default(false).notNull(),
	image: text(),
	role: userRole().default("user").notNull(),
	isOnboarded: boolean("is_onboarded").default(false).notNull(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
	businessId: uuid("business_id").references(() => business.id, { onDelete: "set null" } ),
}, (table) => [
	index("user_business_idx").using("btree", table.businessId.asc().nullsLast()),
	index("user_email_idx").using("btree", table.email.asc().nullsLast()),
	index("user_role_idx").using("btree", table.role.asc().nullsLast()),
	unique("user_email_unique").on(table.email),]);

export const verification = pgTable("verification", {
	id: text().primaryKey(),
	identifier: text().notNull(),
	value: text().notNull(),
	expiresAt: timestamp("expires_at").notNull(),
	createdAt: timestamp("created_at").default(sql`now()`).notNull(),
	updatedAt: timestamp("updated_at").default(sql`now()`).notNull(),
}, (table) => [
	index("verification_identifier_idx").using("btree", table.identifier.asc().nullsLast()),
]);
