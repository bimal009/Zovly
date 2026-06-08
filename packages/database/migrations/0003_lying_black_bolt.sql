ALTER TABLE "business_members" ALTER COLUMN "user_id" SET DATA TYPE text;--> statement-breakpoint
ALTER TABLE "member_invites" ALTER COLUMN "invited_by_id" SET DATA TYPE text;--> statement-breakpoint
ALTER TABLE "member_invites" ALTER COLUMN "revoked_by_id" SET DATA TYPE text;