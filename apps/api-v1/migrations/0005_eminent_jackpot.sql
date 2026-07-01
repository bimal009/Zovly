ALTER TABLE "app_connections" DISABLE ROW LEVEL SECURITY;--> statement-breakpoint
DROP TABLE "app_connections" CASCADE;--> statement-breakpoint
ALTER TABLE "app_credentials" DROP CONSTRAINT "app_cred_app_uq";--> statement-breakpoint
ALTER TABLE "app_credentials" ADD CONSTRAINT "app_cred_account_uq" UNIQUE("business_id","app_name","platform_account_id");