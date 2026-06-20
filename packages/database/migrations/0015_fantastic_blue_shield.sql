CREATE TYPE "public"."message_status" AS ENUM('pending', 'sent', 'failed', 'skipped');--> statement-breakpoint
DROP INDEX "conv_thread_idx";--> statement-breakpoint
ALTER TABLE "conversations" ADD COLUMN "contact_name" text;--> statement-breakpoint
ALTER TABLE "conversations" ADD COLUMN "contact_username" text;--> statement-breakpoint
ALTER TABLE "conversations" ADD COLUMN "contact_avatar_url" text;--> statement-breakpoint
ALTER TABLE "messages" ADD COLUMN "status" "message_status";--> statement-breakpoint
ALTER TABLE "messages" ADD COLUMN "error_message" text;--> statement-breakpoint
ALTER TABLE "messages" ADD COLUMN "sent_to_platform_at" timestamp;--> statement-breakpoint
ALTER TABLE "messages" ADD COLUMN "platform_message_id" text;--> statement-breakpoint
ALTER TABLE "conversations" DROP COLUMN "ai_enabled";--> statement-breakpoint
ALTER TABLE "conversations" ADD CONSTRAINT "conv_thread_uq" UNIQUE("business_id","platform","thread_id");