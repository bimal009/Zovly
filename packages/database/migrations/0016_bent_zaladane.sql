ALTER TABLE "app_credentials" DROP CONSTRAINT "app_cred_account_uq";--> statement-breakpoint
ALTER TABLE "app_credentials" ADD CONSTRAINT "app_cred_app_uq" UNIQUE("business_id","app_name");