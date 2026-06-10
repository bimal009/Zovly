ALTER TABLE "user" DROP CONSTRAINT "user_business_id_business_id_fk";
--> statement-breakpoint
ALTER TABLE "user" ALTER COLUMN "business_id" DROP NOT NULL;--> statement-breakpoint
ALTER TABLE "user" ADD CONSTRAINT "user_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE set null ON UPDATE no action;