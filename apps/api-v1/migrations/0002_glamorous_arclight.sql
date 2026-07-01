ALTER TABLE "user" DROP CONSTRAINT "user_business_id_business_id_fk";
--> statement-breakpoint
DROP INDEX "user_business_idx";--> statement-breakpoint
ALTER TABLE "business_members" ADD COLUMN "can_edit_business" boolean DEFAULT false NOT NULL;--> statement-breakpoint
ALTER TABLE "user" DROP COLUMN "business_id";