ALTER TABLE "business" DROP CONSTRAINT "business_slug_unique";--> statement-breakpoint
DROP INDEX "business_slug_idx";--> statement-breakpoint
ALTER TABLE "business" DROP COLUMN "slug";