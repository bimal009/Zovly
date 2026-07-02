ALTER TABLE "messages" ADD COLUMN "delivered_at" timestamp;--> statement-breakpoint
ALTER TABLE "messages" ADD COLUMN "seen_at" timestamp;--> statement-breakpoint
ALTER TABLE "messages" ADD COLUMN "reply_to_message_id" uuid;--> statement-breakpoint
ALTER TABLE "messages" ADD COLUMN "platform_sender_id" text;--> statement-breakpoint
ALTER TABLE "messages" ADD CONSTRAINT "messages_reply_to_message_id_messages_id_fk" FOREIGN KEY ("reply_to_message_id") REFERENCES "public"."messages"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "msg_platform_id_idx" ON "messages" USING btree ("platform_message_id");--> statement-breakpoint
CREATE INDEX "msg_reply_to_idx" ON "messages" USING btree ("reply_to_message_id");