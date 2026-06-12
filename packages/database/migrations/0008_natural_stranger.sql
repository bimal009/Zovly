ALTER TABLE "products" ADD COLUMN "discount" integer DEFAULT 0 NOT NULL;--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_manage_content";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_view_analytics";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_manage_ads";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_read_dms";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_reply_dms";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_read_comments";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_reply_comments";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_view_leads";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_manage_leads";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_view_bookings";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_manage_bookings";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_view_inventory";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_manage_inventory";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_view_orders";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_manage_settings";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_manage_members";--> statement-breakpoint
ALTER TABLE "member_invites" DROP COLUMN "can_manage_billing";--> statement-breakpoint
ALTER TABLE "products" DROP COLUMN "mrp";--> statement-breakpoint
ALTER TABLE "products" DROP COLUMN "track_inventory";