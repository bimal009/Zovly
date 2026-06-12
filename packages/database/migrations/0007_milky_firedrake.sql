CREATE TYPE "public"."product_status" AS ENUM('active', 'inactive', 'archived');--> statement-breakpoint
CREATE TYPE "public"."billing_interval" AS ENUM('weekly', 'monthly', 'quarterly', 'yearly');--> statement-breakpoint
CREATE TYPE "public"."service_status" AS ENUM('active', 'inactive', 'archived');--> statement-breakpoint
CREATE TYPE "public"."service_type" AS ENUM('appointment', 'membership', 'class', 'package');--> statement-breakpoint
CREATE TABLE "products" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"name" text NOT NULL,
	"description" text,
	"sku" text,
	"status" "product_status" DEFAULT 'active' NOT NULL,
	"price" integer NOT NULL,
	"cost_price" integer,
	"mrp" integer,
	"currency" text DEFAULT 'NPR' NOT NULL,
	"stock_qty" integer DEFAULT 0 NOT NULL,
	"low_stock_threshold" integer DEFAULT 5,
	"track_inventory" boolean DEFAULT true NOT NULL,
	"images" text[] DEFAULT '{}',
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "services" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"business_id" uuid NOT NULL,
	"type" "service_type" DEFAULT 'appointment' NOT NULL,
	"status" "service_status" DEFAULT 'active' NOT NULL,
	"name" text NOT NULL,
	"description" text,
	"price" integer NOT NULL,
	"cost_price" integer,
	"mrp" integer,
	"currency" text DEFAULT 'NPR' NOT NULL,
	"requires_deposit" boolean DEFAULT false NOT NULL,
	"deposit_amount" integer,
	"duration_min" integer,
	"buffer_min" integer DEFAULT 0,
	"max_advance_days" integer DEFAULT 30,
	"google_calendar_id" text,
	"max_concurrent" integer DEFAULT 1,
	"billing_interval" "billing_interval",
	"trial_days" integer DEFAULT 0,
	"session_count" integer,
	"validity_days" integer,
	"images" text[] DEFAULT '{}',
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
ALTER TABLE "products" ADD CONSTRAINT "products_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "services" ADD CONSTRAINT "services_business_id_business_id_fk" FOREIGN KEY ("business_id") REFERENCES "public"."business"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "products_business_id_idx" ON "products" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "products_status_idx" ON "products" USING btree ("status");--> statement-breakpoint
CREATE INDEX "products_sku_idx" ON "products" USING btree ("sku");--> statement-breakpoint
CREATE INDEX "products_business_status_idx" ON "products" USING btree ("business_id","status");--> statement-breakpoint
CREATE INDEX "services_business_id_idx" ON "services" USING btree ("business_id");--> statement-breakpoint
CREATE INDEX "services_type_idx" ON "services" USING btree ("type");--> statement-breakpoint
CREATE INDEX "services_status_idx" ON "services" USING btree ("status");--> statement-breakpoint
CREATE INDEX "services_business_status_idx" ON "services" USING btree ("business_id","status");