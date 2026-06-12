CREATE TYPE "public"."message_direction" AS ENUM('in', 'out');--> statement-breakpoint
CREATE TYPE "public"."message_media_type" AS ENUM('image', 'video', 'audio', 'document');--> statement-breakpoint
CREATE TYPE "public"."message_sender" AS ENUM('ai', 'human');--> statement-breakpoint
CREATE TYPE "public"."platform" AS ENUM('instagram', 'facebook', 'whatsapp', 'tiktok');--> statement-breakpoint
CREATE TABLE "conversations" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"platform" "platform" NOT NULL,
	"thread_id" text NOT NULL,
	"contact_id" text NOT NULL,
	"ai_enabled" boolean DEFAULT true NOT NULL,
	"last_message_at" timestamp DEFAULT now() NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL
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
	"sent_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
ALTER TABLE "services" ADD COLUMN "features" jsonb DEFAULT '[]'::jsonb NOT NULL;--> statement-breakpoint
ALTER TABLE "conversations" ADD CONSTRAINT "conversations_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "messages" ADD CONSTRAINT "messages_conversation_id_conversations_id_fk" FOREIGN KEY ("conversation_id") REFERENCES "public"."conversations"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "messages" ADD CONSTRAINT "messages_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "conv_business_platform_idx" ON "conversations" USING btree ("business_id","platform");--> statement-breakpoint
CREATE INDEX "conv_thread_idx" ON "conversations" USING btree ("business_id","platform","thread_id");--> statement-breakpoint
CREATE INDEX "msg_conversation_idx" ON "messages" USING btree ("conversation_id");--> statement-breakpoint
CREATE INDEX "msg_vectorize_idx" ON "messages" USING btree ("business_id","is_vectorized");