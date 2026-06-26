-- Money columns move from integer minor-units (cents) to numeric(12,2) actual amounts.
-- Existing rows hold cents, so divide by 100 while changing the type.
ALTER TABLE "product_variants" ALTER COLUMN "price" SET DATA TYPE numeric(12, 2) USING ("price" / 100.0);--> statement-breakpoint
ALTER TABLE "product_variants" ALTER COLUMN "cost_price" SET DATA TYPE numeric(12, 2) USING ("cost_price" / 100.0);--> statement-breakpoint
ALTER TABLE "products" ALTER COLUMN "price" SET DATA TYPE numeric(12, 2) USING ("price" / 100.0);--> statement-breakpoint
ALTER TABLE "products" ALTER COLUMN "cost_price" SET DATA TYPE numeric(12, 2) USING ("cost_price" / 100.0);
