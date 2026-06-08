ALTER TABLE "plans" RENAME COLUMN "has_nepal_payments" TO "has_payments";--> statement-breakpoint
ALTER TABLE "plans" DROP COLUMN "has_white_label";--> statement-breakpoint
ALTER TABLE "plans" DROP COLUMN "has_api_access";