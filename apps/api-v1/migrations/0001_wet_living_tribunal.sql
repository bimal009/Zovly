ALTER TABLE "user" ADD COLUMN "business_id" uuid;--> statement-breakpoint
ALTER TABLE "user" ADD CONSTRAINT "user_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "user_business_idx" ON "user" USING btree ("business_id");