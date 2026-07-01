ALTER TABLE "app_credentials" ALTER COLUMN "platform_account_id" SET NOT NULL;--> statement-breakpoint
ALTER TABLE "app_credentials" ADD COLUMN "platform_account_image" text;--> statement-breakpoint
CREATE INDEX "app_cred_active_idx" ON "app_credentials" USING btree ("business_id","app_name","is_active");