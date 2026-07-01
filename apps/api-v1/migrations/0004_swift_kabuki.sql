ALTER TABLE "conversations" DROP CONSTRAINT "conversations_active_product_id_products_id_fk";
--> statement-breakpoint
ALTER TABLE "conversations" DROP COLUMN "active_product_id";--> statement-breakpoint
ALTER TABLE "conversations" DROP COLUMN "active_product_at";