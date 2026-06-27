ALTER TABLE "conversations" ADD COLUMN "active_product_id" uuid;--> statement-breakpoint
ALTER TABLE "conversations" ADD COLUMN "active_product_at" timestamp;--> statement-breakpoint
ALTER TABLE "products" ADD COLUMN "search_tsv" "tsvector" GENERATED ALWAYS AS (to_tsvector('simple', coalesce(name, '') || ' ' || coalesce(description, '')) || array_to_tsvector(array_remove(coalesce(tags, '{}'::text[]), NULL))) STORED;--> statement-breakpoint
ALTER TABLE "conversations" ADD CONSTRAINT "conversations_active_product_id_products_id_fk" FOREIGN KEY ("active_product_id") REFERENCES "public"."products"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "idx_products_search_tsv" ON "products" USING gin ("search_tsv");--> statement-breakpoint
CREATE INDEX "idx_products_name_trgm" ON "products" USING gin ("name" gin_trgm_ops);