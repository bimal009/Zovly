CREATE TYPE "public"."user_role" AS ENUM('vendor', 'admin', 'user');--> statement-breakpoint
CREATE TYPE "public"."business_type" AS ENUM('product', 'service', 'both');--> statement-breakpoint
CREATE TYPE "public"."plan" AS ENUM('starter', 'growth', 'pro', 'agency');--> statement-breakpoint
CREATE TYPE "public"."member_role" AS ENUM('owner', 'admin', 'manager', 'staff', 'viewer');--> statement-breakpoint
CREATE TYPE "public"."permission_action" AS ENUM('read', 'write', 'delete', 'manage');--> statement-breakpoint
CREATE TYPE "public"."billing_interval" AS ENUM('weekly', 'monthly', 'quarterly', 'yearly');--> statement-breakpoint
CREATE TYPE "public"."service_status" AS ENUM('active', 'inactive', 'archived');--> statement-breakpoint
CREATE TYPE "public"."service_type" AS ENUM('appointment', 'membership', 'class', 'package');--> statement-breakpoint
CREATE TYPE "public"."product_status" AS ENUM('active', 'inactive', 'archived');--> statement-breakpoint
CREATE TYPE "public"."knowledge_source_type" AS ENUM('faq', 'policy', 'post', 'product', 'service');--> statement-breakpoint
CREATE TYPE "public"."billing_cycle" AS ENUM('monthly', 'yearly');--> statement-breakpoint
CREATE TYPE "public"."message_direction" AS ENUM('in', 'out');--> statement-breakpoint
CREATE TYPE "public"."message_media_type" AS ENUM('image', 'video', 'audio', 'document', 'link');--> statement-breakpoint
CREATE TYPE "public"."message_sender" AS ENUM('ai', 'human');--> statement-breakpoint
CREATE TYPE "public"."message_status" AS ENUM('pending', 'sent', 'failed', 'skipped');--> statement-breakpoint
CREATE TYPE "public"."payment_method" AS ENUM('esewa', 'khalti', 'fonepay', 'bank_transfer', 'cash', 'other');--> statement-breakpoint
CREATE TYPE "public"."payment_status" AS ENUM('paid', 'refunded', 'partially_refunded', 'failed');--> statement-breakpoint
CREATE TYPE "public"."plan_status" AS ENUM('active', 'trialing', 'past_due', 'paused', 'cancelled', 'expired');--> statement-breakpoint
CREATE TYPE "public"."platform" AS ENUM('instagram', 'facebook', 'whatsapp', 'tiktok');--> statement-breakpoint
CREATE TYPE "public"."invite_status" AS ENUM('pending', 'accepted', 'declined', 'expired', 'revoked');--> statement-breakpoint
CREATE TABLE "account" (
	"id" text PRIMARY KEY NOT NULL,
	"account_id" text NOT NULL,
	"provider_id" text NOT NULL,
	"user_id" text NOT NULL,
	"access_token" text,
	"refresh_token" text,
	"id_token" text,
	"access_token_expires_at" timestamp,
	"refresh_token_expires_at" timestamp,
	"scope" text,
	"password" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp NOT NULL
);
--> statement-breakpoint
CREATE TABLE "session" (
	"id" text PRIMARY KEY NOT NULL,
	"expires_at" timestamp NOT NULL,
	"token" text NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp NOT NULL,
	"ip_address" text,
	"user_agent" text,
	"user_id" text NOT NULL,
	CONSTRAINT "session_token_unique" UNIQUE("token")
);
--> statement-breakpoint
CREATE TABLE "user" (
	"id" text PRIMARY KEY NOT NULL,
	"name" text NOT NULL,
	"email" text NOT NULL,
	"email_verified" boolean DEFAULT false NOT NULL,
	"image" text,
	"role" "user_role" DEFAULT 'user' NOT NULL,
	"is_onboarded" boolean DEFAULT false NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "user_email_unique" UNIQUE("email")
);
--> statement-breakpoint
CREATE TABLE "verification" (
	"id" text PRIMARY KEY NOT NULL,
	"identifier" text NOT NULL,
	"value" text NOT NULL,
	"expires_at" timestamp NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "business" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"name" text NOT NULL,
	"description" text NOT NULL,
	"logo" text,
	"website" text,
	"phone" text,
	"address" text NOT NULL,
	"city" text,
	"country" text DEFAULT 'NPL',
	"type" "business_type" DEFAULT 'service' NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "app_connections" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"instagram" boolean DEFAULT false NOT NULL,
	"facebook" boolean DEFAULT false NOT NULL,
	"tiktok" boolean DEFAULT false NOT NULL,
	"whatsapp" boolean DEFAULT false NOT NULL,
	"google_workspace" boolean DEFAULT false NOT NULL,
	"stripe_connect" boolean DEFAULT false NOT NULL,
	"fonepay" boolean DEFAULT false NOT NULL,
	"khalti" boolean DEFAULT false NOT NULL,
	"esewa" boolean DEFAULT false NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "app_connections_business_id_unique" UNIQUE("business_id")
);
--> statement-breakpoint
CREATE TABLE "business_members" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"user_id" text NOT NULL,
	"role" "member_role" DEFAULT 'viewer' NOT NULL,
	"can_manage_content" boolean DEFAULT false NOT NULL,
	"can_view_analytics" boolean DEFAULT false NOT NULL,
	"can_manage_ads" boolean DEFAULT false NOT NULL,
	"can_read_dms" boolean DEFAULT false NOT NULL,
	"can_reply_dms" boolean DEFAULT false NOT NULL,
	"can_read_comments" boolean DEFAULT false NOT NULL,
	"can_reply_comments" boolean DEFAULT false NOT NULL,
	"can_view_leads" boolean DEFAULT false NOT NULL,
	"can_manage_leads" boolean DEFAULT false NOT NULL,
	"can_view_bookings" boolean DEFAULT false NOT NULL,
	"can_manage_bookings" boolean DEFAULT false NOT NULL,
	"can_view_inventory" boolean DEFAULT false NOT NULL,
	"can_manage_inventory" boolean DEFAULT false NOT NULL,
	"can_view_orders" boolean DEFAULT false NOT NULL,
	"can_manage_settings" boolean DEFAULT false NOT NULL,
	"can_manage_members" boolean DEFAULT false NOT NULL,
	"can_manage_billing" boolean DEFAULT false NOT NULL,
	"joined_at" timestamp DEFAULT now() NOT NULL,
	"last_seen_at" timestamp,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "uq_business_member" UNIQUE("business_id","user_id")
);
--> statement-breakpoint
CREATE TABLE "plans" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"name" text NOT NULL,
	"paddle_product_id" text,
	"paddle_price_id_monthly" text,
	"paddle_price_id_yearly" text,
	"monthly_price" integer NOT NULL,
	"yearly_price" integer NOT NULL,
	"max_members" integer NOT NULL,
	"max_social_accounts" integer NOT NULL,
	"max_ai_replies_month" integer NOT NULL,
	"max_posts_month" integer NOT NULL,
	"max_leads" integer NOT NULL,
	"max_products" integer NOT NULL,
	"max_bookings_month" integer NOT NULL,
	"has_video_upload" boolean DEFAULT false NOT NULL,
	"has_multi_platform_post" boolean DEFAULT false NOT NULL,
	"has_post_analytics" boolean DEFAULT false NOT NULL,
	"has_ai_dm_replies" boolean DEFAULT false NOT NULL,
	"has_ai_comment_replies" boolean DEFAULT false NOT NULL,
	"has_ai_lead_scoring" boolean DEFAULT false NOT NULL,
	"has_ai_ad_suggestions" boolean DEFAULT false NOT NULL,
	"has_voice_transcription" boolean DEFAULT false NOT NULL,
	"has_image_understanding" boolean DEFAULT false NOT NULL,
	"has_bookings" boolean DEFAULT false NOT NULL,
	"has_inventory" boolean DEFAULT false NOT NULL,
	"has_payments" boolean DEFAULT false NOT NULL,
	"has_google_workspace" boolean DEFAULT false NOT NULL,
	"has_meta_ads" boolean DEFAULT false NOT NULL,
	"has_tiktok_ads" boolean DEFAULT false NOT NULL,
	"has_priority_support" boolean DEFAULT false NOT NULL,
	"ai_reply_overage_price_usd_per_500" integer,
	"is_active" boolean DEFAULT true NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "plans_name_unique" UNIQUE("name"),
	CONSTRAINT "plans_paddle_product_id_unique" UNIQUE("paddle_product_id"),
	CONSTRAINT "plans_paddle_price_id_monthly_unique" UNIQUE("paddle_price_id_monthly"),
	CONSTRAINT "plans_paddle_price_id_yearly_unique" UNIQUE("paddle_price_id_yearly")
);
--> statement-breakpoint
CREATE TABLE "business_subscriptions" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"plan_id" uuid NOT NULL,
	"paddle_subscription_id" text,
	"paddle_customer_id" text,
	"paddle_price_id" text,
	"billing_cycle" "billing_cycle" DEFAULT 'monthly' NOT NULL,
	"status" "plan_status" DEFAULT 'trialing' NOT NULL,
	"ai_replies_used" integer DEFAULT 0 NOT NULL,
	"posts_used" integer DEFAULT 0 NOT NULL,
	"usage_reset_at" timestamp,
	"trial_started_at" timestamp,
	"trial_ends_at" timestamp,
	"current_period_start" timestamp,
	"current_period_end" timestamp,
	"cancel_at_period_end" boolean DEFAULT false NOT NULL,
	"cancelled_at" timestamp,
	"paused_at" timestamp,
	"notes" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "business_subscriptions_business_id_unique" UNIQUE("business_id"),
	CONSTRAINT "business_subscriptions_paddle_subscription_id_unique" UNIQUE("paddle_subscription_id")
);
--> statement-breakpoint
CREATE TABLE "payment_records" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"subscription_id" uuid,
	"plan_id" uuid,
	"billing_cycle" "billing_cycle" NOT NULL,
	"paddle_transaction_id" text NOT NULL,
	"paddle_subscription_id" text,
	"paddle_customer_id" text,
	"amount" integer NOT NULL,
	"currency" text DEFAULT 'USD' NOT NULL,
	"period_start" timestamp NOT NULL,
	"period_end" timestamp NOT NULL,
	"status" "payment_status" DEFAULT 'paid' NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "payment_records_paddle_transaction_id_unique" UNIQUE("paddle_transaction_id")
);
--> statement-breakpoint
CREATE TABLE "services" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"type" "service_type" DEFAULT 'appointment' NOT NULL,
	"status" "service_status" DEFAULT 'active' NOT NULL,
	"name" text NOT NULL,
	"description" text,
	"price" integer NOT NULL,
	"cost_price" integer,
	"mrp" integer,
	"currency" text DEFAULT 'NPR' NOT NULL,
	"requires_deposit" boolean DEFAULT false NOT NULL,
	"deposit_amount" integer,
	"duration_min" integer,
	"buffer_min" integer DEFAULT 0,
	"max_advance_days" integer DEFAULT 30,
	"google_calendar_id" text,
	"max_concurrent" integer DEFAULT 1,
	"billing_interval" "billing_interval",
	"trial_days" integer DEFAULT 0,
	"session_count" integer,
	"validity_days" integer,
	"features" jsonb DEFAULT '[]'::jsonb NOT NULL,
	"images" text[] DEFAULT '{}',
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "products" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"category_id" uuid,
	"name" text NOT NULL,
	"description" text,
	"sku" text,
	"status" "product_status" DEFAULT 'active' NOT NULL,
	"tags" text[] DEFAULT '{}',
	"attributes" jsonb,
	"price" numeric(12, 2) NOT NULL,
	"cost_price" numeric(12, 2),
	"discount" integer DEFAULT 0 NOT NULL,
	"currency" text DEFAULT 'NPR' NOT NULL,
	"stock_qty" integer DEFAULT 0 NOT NULL,
	"low_stock_threshold" integer DEFAULT 5,
	"images" text[] DEFAULT '{}',
	"search_tsv" "tsvector" GENERATED ALWAYS AS (to_tsvector('simple', coalesce(name, '') || ' ' || coalesce(description, '')) || array_to_tsvector(array_remove(coalesce(tags, '{}'::text[]), NULL))) STORED,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "faqs" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"question" text NOT NULL,
	"answer" text NOT NULL,
	"is_active" boolean DEFAULT true NOT NULL,
	"sort_order" integer DEFAULT 0 NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "knowledge_chunks" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"source_type" "knowledge_source_type" NOT NULL,
	"source_id" uuid NOT NULL,
	"chunk_index" integer DEFAULT 0 NOT NULL,
	"content" text NOT NULL,
	"embedding" vector(1024) NOT NULL,
	"metadata" jsonb,
	"created_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "kc_source_chunk_uq" UNIQUE("source_type","source_id","chunk_index")
);
--> statement-breakpoint
CREATE TABLE "policies" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"title" text NOT NULL,
	"content" text NOT NULL,
	"is_active" boolean DEFAULT true NOT NULL,
	"sort_order" integer DEFAULT 0 NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "conversations" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"platform" "platform" NOT NULL,
	"thread_id" text NOT NULL,
	"contact_id" text NOT NULL,
	"contact_name" text,
	"contact_username" text,
	"contact_avatar_url" text,
	"last_message_at" timestamp DEFAULT now() NOT NULL,
	"active_product_id" uuid,
	"active_product_at" timestamp,
	"created_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "conv_thread_uq" UNIQUE("business_id","platform","thread_id")
);
--> statement-breakpoint
CREATE TABLE "messages" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"conversation_id" uuid NOT NULL,
	"business_id" uuid NOT NULL,
	"direction" "message_direction" NOT NULL,
	"sent_by" "message_sender",
	"content" text,
	"media_url" text,
	"media_type" "message_media_type",
	"is_vectorized" boolean DEFAULT false NOT NULL,
	"status" "message_status",
	"error_message" text,
	"sent_to_platform_at" timestamp,
	"platform_message_id" text,
	"sent_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "message_embeddings" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"message_id" uuid NOT NULL,
	"business_id" uuid NOT NULL,
	"conversation_id" uuid NOT NULL,
	"content" text NOT NULL,
	"embedding" vector(1024) NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "message_embeddings_message_id_unique" UNIQUE("message_id")
);
--> statement-breakpoint
CREATE TABLE "member_invites" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"invited_by_id" text,
	"invited_email" text NOT NULL,
	"role" "member_role" DEFAULT 'viewer' NOT NULL,
	"token" text NOT NULL,
	"status" "invite_status" DEFAULT 'pending' NOT NULL,
	"expires_at" timestamp NOT NULL,
	"accepted_at" timestamp,
	"declined_at" timestamp,
	"revoked_at" timestamp,
	"revoked_by_id" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "member_invites_token_unique" UNIQUE("token")
);
--> statement-breakpoint
CREATE TABLE "app_credentials" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"app_name" text NOT NULL,
	"access_token" text,
	"refresh_token" text,
	"token_expires_at" timestamp,
	"scopes" text[],
	"public_key" text,
	"secret_key" text,
	"merchant_id" text,
	"platform_account_id" text,
	"platform_account_name" text,
	"webhook_verify_token" text,
	"webhook_subscribed_at" timestamp,
	"is_active" boolean DEFAULT true NOT NULL,
	"connected_at" timestamp,
	"disconnected_at" timestamp,
	"last_sync_at" timestamp,
	"error_message" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "app_cred_app_uq" UNIQUE("business_id","app_name")
);
--> statement-breakpoint
CREATE TABLE "categories" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"name" text NOT NULL,
	"description" text,
	"slug" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "categories_business_name_uq" UNIQUE("business_id","name")
);
--> statement-breakpoint
CREATE TABLE "product_variants" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"product_id" uuid NOT NULL,
	"business_id" uuid NOT NULL,
	"name" text NOT NULL,
	"sku" text,
	"attributes" jsonb,
	"price" numeric(12, 2),
	"cost_price" numeric(12, 2),
	"discount" integer,
	"stock_qty" integer DEFAULT 0 NOT NULL,
	"low_stock_threshold" integer DEFAULT 5,
	"images" text[] DEFAULT '{}',
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "pv_product_name_uq" UNIQUE("product_id","name")
);
--> statement-breakpoint
ALTER TABLE "account" ADD CONSTRAINT "account_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "session" ADD CONSTRAINT "session_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "app_connections" ADD CONSTRAINT "app_connections_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "business_members" ADD CONSTRAINT "business_members_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "business_members" ADD CONSTRAINT "business_members_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "business_subscriptions" ADD CONSTRAINT "business_subscriptions_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "business_subscriptions" ADD CONSTRAINT "business_subscriptions_plan_id_plans_id_fk" FOREIGN KEY ("plan_id") REFERENCES "public"."plans"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "payment_records" ADD CONSTRAINT "payment_records_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "payment_records" ADD CONSTRAINT "payment_records_subscription_id_business_subscriptions_id_fk" FOREIGN KEY ("subscription_id") REFERENCES "public"."business_subscriptions"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "payment_records" ADD CONSTRAINT "payment_records_plan_id_plans_id_fk" FOREIGN KEY ("plan_id") REFERENCES "public"."plans"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "services" ADD CONSTRAINT "services_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "products" ADD CONSTRAINT "products_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "products" ADD CONSTRAINT "products_category_id_categories_id_fk" FOREIGN KEY ("category_id") REFERENCES "public"."categories"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "faqs" ADD CONSTRAINT "faqs_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "knowledge_chunks" ADD CONSTRAINT "knowledge_chunks_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "policies" ADD CONSTRAINT "policies_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "conversations" ADD CONSTRAINT "conversations_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "conversations" ADD CONSTRAINT "conversations_active_product_id_products_id_fk" FOREIGN KEY ("active_product_id") REFERENCES "public"."products"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "messages" ADD CONSTRAINT "messages_conversation_id_conversations_id_fk" FOREIGN KEY ("conversation_id") REFERENCES "public"."conversations"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "messages" ADD CONSTRAINT "messages_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "message_embeddings" ADD CONSTRAINT "message_embeddings_message_id_messages_id_fk" FOREIGN KEY ("message_id") REFERENCES "public"."messages"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "message_embeddings" ADD CONSTRAINT "message_embeddings_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "message_embeddings" ADD CONSTRAINT "message_embeddings_conversation_id_conversations_id_fk" FOREIGN KEY ("conversation_id") REFERENCES "public"."conversations"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "member_invites" ADD CONSTRAINT "member_invites_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "member_invites" ADD CONSTRAINT "member_invites_invited_by_id_user_id_fk" FOREIGN KEY ("invited_by_id") REFERENCES "public"."user"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "member_invites" ADD CONSTRAINT "member_invites_revoked_by_id_user_id_fk" FOREIGN KEY ("revoked_by_id") REFERENCES "public"."user"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "app_credentials" ADD CONSTRAINT "app_credentials_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "categories" ADD CONSTRAINT "categories_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "product_variants" ADD CONSTRAINT "product_variants_product_id_products_id_fk" FOREIGN KEY ("product_id") REFERENCES "public"."products"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "product_variants" ADD CONSTRAINT "product_variants_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "account_user_idx" ON "account" USING btree ("user_id");--> statement-breakpoint
CREATE INDEX "account_provider_idx" ON "account" USING btree ("provider_id");--> statement-breakpoint
CREATE INDEX "account_accountId_idx" ON "account" USING btree ("account_id");--> statement-breakpoint
CREATE INDEX "session_userId_idx" ON "session" USING btree ("user_id");--> statement-breakpoint
CREATE INDEX "user_email_idx" ON "user" USING btree ("email");--> statement-breakpoint
CREATE INDEX "user_role_idx" ON "user" USING btree ("role");--> statement-breakpoint
CREATE INDEX "verification_identifier_idx" ON "verification" USING btree ("identifier");--> statement-breakpoint
CREATE INDEX "business_type_idx" ON "business" USING btree ("type");--> statement-breakpoint
CREATE INDEX "app_conn_business_idx" ON "app_connections" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "biz_member_business_idx" ON "business_members" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "biz_member_user_idx" ON "business_members" USING btree ("user_id");--> statement-breakpoint
CREATE INDEX "biz_member_role_idx" ON "business_members" USING btree ("role");--> statement-breakpoint
CREATE INDEX "plan_name_idx" ON "plans" USING btree ("name");--> statement-breakpoint
CREATE INDEX "plan_active_idx" ON "plans" USING btree ("is_active");--> statement-breakpoint
CREATE INDEX "plan_paddle_product_idx" ON "plans" USING btree ("paddle_product_id");--> statement-breakpoint
CREATE INDEX "plan_paddle_monthly_idx" ON "plans" USING btree ("paddle_price_id_monthly");--> statement-breakpoint
CREATE INDEX "plan_paddle_yearly_idx" ON "plans" USING btree ("paddle_price_id_yearly");--> statement-breakpoint
CREATE INDEX "sub_business_idx" ON "business_subscriptions" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "sub_plan_idx" ON "business_subscriptions" USING btree ("plan_id");--> statement-breakpoint
CREATE INDEX "sub_status_idx" ON "business_subscriptions" USING btree ("status");--> statement-breakpoint
CREATE INDEX "sub_paddle_sub_idx" ON "business_subscriptions" USING btree ("paddle_subscription_id");--> statement-breakpoint
CREATE INDEX "sub_paddle_customer_idx" ON "business_subscriptions" USING btree ("paddle_customer_id");--> statement-breakpoint
CREATE INDEX "payment_business_idx" ON "payment_records" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "payment_status_idx" ON "payment_records" USING btree ("status");--> statement-breakpoint
CREATE INDEX "payment_paddle_txn_idx" ON "payment_records" USING btree ("paddle_transaction_id");--> statement-breakpoint
CREATE INDEX "payment_paddle_sub_idx" ON "payment_records" USING btree ("paddle_subscription_id");--> statement-breakpoint
CREATE INDEX "payment_sub_idx" ON "payment_records" USING btree ("subscription_id");--> statement-breakpoint
CREATE INDEX "services_business_id_idx" ON "services" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "services_type_idx" ON "services" USING btree ("type");--> statement-breakpoint
CREATE INDEX "services_status_idx" ON "services" USING btree ("status");--> statement-breakpoint
CREATE INDEX "services_business_status_idx" ON "services" USING btree ("business_id","status");--> statement-breakpoint
CREATE INDEX "products_business_id_idx" ON "products" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "products_category_id_idx" ON "products" USING btree ("category_id");--> statement-breakpoint
CREATE INDEX "products_status_idx" ON "products" USING btree ("status");--> statement-breakpoint
CREATE INDEX "products_sku_idx" ON "products" USING btree ("sku");--> statement-breakpoint
CREATE INDEX "products_business_status_idx" ON "products" USING btree ("business_id","status");--> statement-breakpoint
CREATE INDEX "idx_products_search_tsv" ON "products" USING gin ("search_tsv");--> statement-breakpoint
CREATE INDEX "idx_products_name_trgm" ON "products" USING gin ("name" gin_trgm_ops);--> statement-breakpoint
CREATE INDEX "faq_business_idx" ON "faqs" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "kc_hnsw_idx" ON "knowledge_chunks" USING hnsw ("embedding" vector_cosine_ops);--> statement-breakpoint
CREATE INDEX "kc_business_idx" ON "knowledge_chunks" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "kc_source_idx" ON "knowledge_chunks" USING btree ("business_id","source_type","source_id");--> statement-breakpoint
CREATE INDEX "policy_business_idx" ON "policies" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "conv_business_platform_idx" ON "conversations" USING btree ("business_id","platform");--> statement-breakpoint
CREATE INDEX "msg_conversation_idx" ON "messages" USING btree ("conversation_id");--> statement-breakpoint
CREATE INDEX "msg_vectorize_idx" ON "messages" USING btree ("business_id","is_vectorized");--> statement-breakpoint
CREATE INDEX "msg_emb_hnsw_idx" ON "message_embeddings" USING hnsw ("embedding" vector_cosine_ops);--> statement-breakpoint
CREATE INDEX "msg_emb_business_idx" ON "message_embeddings" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "msg_emb_conversation_idx" ON "message_embeddings" USING btree ("conversation_id");--> statement-breakpoint
CREATE INDEX "invite_business_idx" ON "member_invites" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "invite_email_idx" ON "member_invites" USING btree ("invited_email");--> statement-breakpoint
CREATE INDEX "invite_token_idx" ON "member_invites" USING btree ("token");--> statement-breakpoint
CREATE INDEX "invite_status_idx" ON "member_invites" USING btree ("status");--> statement-breakpoint
CREATE INDEX "app_cred_business_idx" ON "app_credentials" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "app_cred_app_name_idx" ON "app_credentials" USING btree ("app_name");--> statement-breakpoint
CREATE INDEX "categories_business_idx" ON "categories" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "pv_product_idx" ON "product_variants" USING btree ("product_id");--> statement-breakpoint
CREATE INDEX "pv_business_idx" ON "product_variants" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "pv_sku_idx" ON "product_variants" USING btree ("sku");