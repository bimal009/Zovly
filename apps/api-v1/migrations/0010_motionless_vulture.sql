ALTER TYPE "public"."platform" ADD VALUE 'web';--> statement-breakpoint
CREATE UNIQUE INDEX "convo_business_thread_idx" ON "conversations" USING btree ("business_id","thread_id");