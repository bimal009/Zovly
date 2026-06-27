ALTER TABLE "conversations" ADD COLUMN "active_product_id" uuid;--> statement-breakpoint
ALTER TABLE "conversations" ADD COLUMN "active_product_at" timestamp;--> statement-breakpoint
ALTER TABLE "conversations" ADD CONSTRAINT "conversations_active_product_id_products_id_fk" FOREIGN KEY ("active_product_id") REFERENCES "public"."products"("id") ON DELETE set null ON UPDATE no action;