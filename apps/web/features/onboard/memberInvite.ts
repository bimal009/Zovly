import type { BusinessMember, MemberRole } from "@repo/database/schema";

type PermissionKey =
  | "canManageContent"
  | "canViewAnalytics"
  | "canManageAds"
  | "canReadDms"
  | "canReplyDms"
  | "canReadComments"
  | "canReplyComments"
  | "canViewLeads"
  | "canManageLeads"
  | "canViewBookings"
  | "canManageBookings"
  | "canViewInventory"
  | "canManageInventory"
  | "canViewOrders"
  | "canManageSettings"
  | "canManageMembers"
  | "canManageBilling";

export const ROLE_DEFAULTS: Record<
  MemberRole,
  Record<PermissionKey, boolean>
> = {
  owner: {
    canManageContent: true,
    canViewAnalytics: true,
    canManageAds: true,
    canReadDms: true,
    canReplyDms: true,
    canReadComments: true,
    canReplyComments: true,
    canViewLeads: true,
    canManageLeads: true,
    canViewBookings: true,
    canManageBookings: true,
    canViewInventory: true,
    canManageInventory: true,
    canViewOrders: true,
    canManageSettings: true,
    canManageMembers: true,
    canManageBilling: true,
  },
  admin: {
    canManageContent: true,
    canViewAnalytics: true,
    canManageAds: true,
    canReadDms: true,
    canReplyDms: true,
    canReadComments: true,
    canReplyComments: true,
    canViewLeads: true,
    canManageLeads: true,
    canViewBookings: true,
    canManageBookings: true,
    canViewInventory: true,
    canManageInventory: true,
    canViewOrders: true,
    canManageSettings: true,
    canManageMembers: true,
    canManageBilling: false, // ← no billing
  },
  manager: {
    canManageContent: true,
    canViewAnalytics: true,
    canManageAds: true,
    canReadDms: true,
    canReplyDms: true,
    canReadComments: true,
    canReplyComments: true,
    canViewLeads: true,
    canManageLeads: true,
    canViewBookings: true,
    canManageBookings: true,
    canViewInventory: true,
    canManageInventory: true,
    canViewOrders: true,
    canManageSettings: false,
    canManageMembers: false,
    canManageBilling: false,
  },
  staff: {
    canManageContent: false,
    canViewAnalytics: false,
    canManageAds: false,
    canReadDms: true,
    canReplyDms: true,
    canReadComments: true,
    canReplyComments: true,
    canViewLeads: false,
    canManageLeads: false,
    canViewBookings: true,
    canManageBookings: false,
    canViewInventory: false,
    canManageInventory: false,
    canViewOrders: false,
    canManageSettings: false,
    canManageMembers: false,
    canManageBilling: false,
  },
  viewer: {
    canManageContent: false,
    canViewAnalytics: true,
    canManageAds: false,
    canReadDms: false,
    canReplyDms: false,
    canReadComments: false,
    canReplyComments: false,
    canViewLeads: true,
    canManageLeads: false,
    canViewBookings: true,
    canManageBookings: false,
    canViewInventory: true,
    canManageInventory: false,
    canViewOrders: true,
    canManageSettings: false,
    canManageMembers: false,
    canManageBilling: false,
  },
};

// Resolves final permission — override wins, falls back to role default
export function resolvePermission(
  member: Pick<BusinessMember, "role"> &
    Partial<Record<PermissionKey, boolean | null>>,
  permission: PermissionKey,
): boolean {
  const override = member[permission];
  if (override !== null && override !== undefined) return override;
  return ROLE_DEFAULTS[member.role][permission];
}
