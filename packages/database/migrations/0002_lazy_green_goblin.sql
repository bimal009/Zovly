CREATE TYPE "public"."business_type" AS ENUM('product', 'service', 'both');--> statement-breakpoint
CREATE TYPE "public"."plan" AS ENUM('starter', 'growth', 'pro', 'agency');--> statement-breakpoint
CREATE TYPE "public"."invite_status" AS ENUM('pending', 'accepted', 'declined', 'expired', 'revoked');--> statement-breakpoint
CREATE TYPE "public"."member_role" AS ENUM('owner', 'admin', 'manager', 'staff', 'viewer');--> statement-breakpoint
CREATE TYPE "public"."permission_action" AS ENUM('read', 'write', 'delete', 'manage');--> statement-breakpoint
CREATE TYPE "public"."billing_cycle" AS ENUM('monthly', 'yearly');--> statement-breakpoint
CREATE TYPE "public"."payment_method" AS ENUM('esewa', 'khalti', 'fonepay', 'bank_transfer', 'cash', 'other');--> statement-breakpoint
CREATE TYPE "public"."payment_status" AS ENUM('pending', 'confirmed', 'rejected');--> statement-breakpoint
CREATE TYPE "public"."plan_status" AS ENUM('active', 'trialing', 'past_due', 'cancelled', 'expired');--> statement-breakpoint
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
	"connected_at" timestamp,
	"disconnected_at" timestamp,
	"last_sync_at" timestamp,
	"error_message" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "business" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"name" text NOT NULL,
	"slug" text NOT NULL,
	"description" text,
	"logo" text,
	"website" text,
	"phone" text,
	"address" text,
	"city" text,
	"country" text DEFAULT 'NP',
	"type" "business_type" DEFAULT 'service' NOT NULL,
	"plan" "plan" DEFAULT 'starter' NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "business_slug_unique" UNIQUE("slug")
);
--> statement-breakpoint
CREATE TABLE "business_members" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"user_id" uuid NOT NULL,
	"role" "member_role" DEFAULT 'viewer' NOT NULL,
	"can_manage_content" boolean,
	"can_view_analytics" boolean,
	"can_manage_ads" boolean,
	"can_read_dms" boolean,
	"can_reply_dms" boolean,
	"can_read_comments" boolean,
	"can_reply_comments" boolean,
	"can_view_leads" boolean,
	"can_manage_leads" boolean,
	"can_view_bookings" boolean,
	"can_manage_bookings" boolean,
	"can_view_inventory" boolean,
	"can_manage_inventory" boolean,
	"can_view_orders" boolean,
	"can_manage_settings" boolean,
	"can_manage_members" boolean,
	"can_manage_billing" boolean,
	"joined_at" timestamp DEFAULT now() NOT NULL,
	"last_seen_at" timestamp,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "uq_business_member" UNIQUE("business_id","user_id")
);
--> statement-breakpoint
CREATE TABLE "member_invites" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"invited_by_id" uuid,
	"invited_email" text NOT NULL,
	"role" "member_role" DEFAULT 'viewer' NOT NULL,
	"can_manage_content" boolean,
	"can_view_analytics" boolean,
	"can_manage_ads" boolean,
	"can_read_dms" boolean,
	"can_reply_dms" boolean,
	"can_read_comments" boolean,
	"can_reply_comments" boolean,
	"can_view_leads" boolean,
	"can_manage_leads" boolean,
	"can_view_bookings" boolean,
	"can_manage_bookings" boolean,
	"can_view_inventory" boolean,
	"can_manage_inventory" boolean,
	"can_view_orders" boolean,
	"can_manage_settings" boolean,
	"can_manage_members" boolean,
	"can_manage_billing" boolean,
	"token" text NOT NULL,
	"status" "invite_status" DEFAULT 'pending' NOT NULL,
	"expires_at" timestamp NOT NULL,
	"accepted_at" timestamp,
	"declined_at" timestamp,
	"revoked_at" timestamp,
	"revoked_by_id" uuid,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "member_invites_token_unique" UNIQUE("token")
);
--> statement-breakpoint
CREATE TABLE "business_subscriptions" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"plan_id" uuid NOT NULL,
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
	"notes" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "business_subscriptions_business_id_unique" UNIQUE("business_id")
);
--> statement-breakpoint
CREATE TABLE "payment_records" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"subscription_id" uuid,
	"plan_id" uuid,
	"billing_cycle" "billing_cycle" NOT NULL,
	"amount_usd" integer NOT NULL,
	"period_start" timestamp NOT NULL,
	"period_end" timestamp NOT NULL,
	"payment_method" "payment_method" NOT NULL,
	"status" "payment_status" DEFAULT 'pending' NOT NULL,
	"screenshot_url" text,
	"txn_reference" text,
	"confirmed_at" timestamp,
	"rejected_at" timestamp,
	"rejected_note" text,
	"notes" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "plans" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"name" text NOT NULL,
	"monthly_price_usd" integer NOT NULL,
	"yearly_price_usd" integer NOT NULL,
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
	"has_nepal_payments" boolean DEFAULT false NOT NULL,
	"has_google_workspace" boolean DEFAULT false NOT NULL,
	"has_meta_ads" boolean DEFAULT false NOT NULL,
	"has_tiktok_ads" boolean DEFAULT false NOT NULL,
	"has_white_label" boolean DEFAULT false NOT NULL,
	"has_api_access" boolean DEFAULT false NOT NULL,
	"has_priority_support" boolean DEFAULT false NOT NULL,
	"ai_reply_overage_price_usd_per_500" integer,
	"is_active" boolean DEFAULT true NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "plans_name_unique" UNIQUE("name")
);
--> statement-breakpoint
ALTER TABLE "user" ADD COLUMN "business_id" uuid NOT NULL;--> statement-breakpoint
ALTER TABLE "app_connections" ADD CONSTRAINT "app_connections_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "app_credentials" ADD CONSTRAINT "app_credentials_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "business_members" ADD CONSTRAINT "business_members_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "business_members" ADD CONSTRAINT "business_members_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "member_invites" ADD CONSTRAINT "member_invites_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "member_invites" ADD CONSTRAINT "member_invites_invited_by_id_user_id_fk" FOREIGN KEY ("invited_by_id") REFERENCES "public"."user"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "member_invites" ADD CONSTRAINT "member_invites_revoked_by_id_user_id_fk" FOREIGN KEY ("revoked_by_id") REFERENCES "public"."user"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "business_subscriptions" ADD CONSTRAINT "business_subscriptions_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "business_subscriptions" ADD CONSTRAINT "business_subscriptions_plan_id_plans_id_fk" FOREIGN KEY ("plan_id") REFERENCES "public"."plans"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "payment_records" ADD CONSTRAINT "payment_records_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "payment_records" ADD CONSTRAINT "payment_records_subscription_id_business_subscriptions_id_fk" FOREIGN KEY ("subscription_id") REFERENCES "public"."business_subscriptions"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "payment_records" ADD CONSTRAINT "payment_records_plan_id_plans_id_fk" FOREIGN KEY ("plan_id") REFERENCES "public"."plans"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "app_conn_business_idx" ON "app_connections" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "app_cred_business_idx" ON "app_credentials" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "app_cred_app_name_idx" ON "app_credentials" USING btree ("app_name");--> statement-breakpoint
CREATE INDEX "business_slug_idx" ON "business" USING btree ("slug");--> statement-breakpoint
CREATE INDEX "business_type_idx" ON "business" USING btree ("type");--> statement-breakpoint
CREATE INDEX "business_plan_idx" ON "business" USING btree ("plan");--> statement-breakpoint
CREATE INDEX "biz_member_business_idx" ON "business_members" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "biz_member_user_idx" ON "business_members" USING btree ("user_id");--> statement-breakpoint
CREATE INDEX "biz_member_role_idx" ON "business_members" USING btree ("role");--> statement-breakpoint
CREATE INDEX "invite_business_idx" ON "member_invites" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "invite_email_idx" ON "member_invites" USING btree ("invited_email");--> statement-breakpoint
CREATE INDEX "invite_token_idx" ON "member_invites" USING btree ("token");--> statement-breakpoint
CREATE INDEX "invite_status_idx" ON "member_invites" USING btree ("status");--> statement-breakpoint
CREATE INDEX "sub_business_idx" ON "business_subscriptions" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "sub_plan_idx" ON "business_subscriptions" USING btree ("plan_id");--> statement-breakpoint
CREATE INDEX "sub_status_idx" ON "business_subscriptions" USING btree ("status");--> statement-breakpoint
CREATE INDEX "payment_business_idx" ON "payment_records" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "payment_status_idx" ON "payment_records" USING btree ("status");--> statement-breakpoint
CREATE INDEX "payment_method_idx" ON "payment_records" USING btree ("payment_method");--> statement-breakpoint
CREATE INDEX "payment_sub_idx" ON "payment_records" USING btree ("subscription_id");--> statement-breakpoint
CREATE INDEX "plan_name_idx" ON "plans" USING btree ("name");--> statement-breakpoint
CREATE INDEX "plan_active_idx" ON "plans" USING btree ("is_active");--> statement-breakpoint
ALTER TABLE "user" ADD CONSTRAINT "user_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "user_business_idx" ON "user" USING btree ("business_id");