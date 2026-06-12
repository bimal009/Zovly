"use client";

import { Clock, MoreHorizontal, Users, ImageOff } from "lucide-react";
import { Badge } from "@repo/ui/components/ui/badge";
import { Button } from "@repo/ui/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@repo/ui/components/ui/dropdown-menu";
import type { ColumnDef } from "@repo/ui/components/ui/data-table";
import type { Service, ServiceStatus, ServiceType } from "../api/services";

// ─── Shared config (also used by services-table stats) ───────────────────────

export const STATUS_CONFIG: Record<
  ServiceStatus,
  { label: string; variant: "default" | "secondary" | "outline" }
> = {
  active: { label: "Active", variant: "default" },
  inactive: { label: "Inactive", variant: "secondary" },
  archived: { label: "Archived", variant: "outline" },
};

export const TYPE_CONFIG: Record<
  ServiceType,
  { label: string; className: string }
> = {
  appointment: {
    label: "Appointment",
    className:
      "bg-blue-100 text-blue-700 border-blue-200 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-800",
  },
  class: {
    label: "Class",
    className:
      "bg-purple-100 text-purple-700 border-purple-200 dark:bg-purple-950 dark:text-purple-300 dark:border-purple-800",
  },
  membership: {
    label: "Membership",
    className:
      "bg-orange-100 text-orange-700 border-orange-200 dark:bg-orange-950 dark:text-orange-300 dark:border-orange-800",
  },
  package: {
    label: "Package",
    className:
      "bg-green-100 text-green-700 border-green-200 dark:bg-green-950 dark:text-green-300 dark:border-green-800",
  },
};

export function formatDuration(min: number | null): string {
  if (min == null) return "—";
  if (min >= 60) {
    const h = Math.floor(min / 60);
    const m = min % 60;
    return m > 0 ? `${h}h ${m}m` : `${h}h`;
  }
  return `${min}m`;
}

function formatPrice(cents: number, currency: string) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    minimumFractionDigits: 2,
  }).format(cents / 100);
}

// ─── Column factory ───────────────────────────────────────────────────────────

interface ColumnActions {
  onEdit: (s: Service) => void;
  onDelete: (s: Service) => void;
  onToggleStatus: (s: Service) => void;
}

export function getServiceColumns({
  onEdit,
  onDelete,
  onToggleStatus,
}: ColumnActions): ColumnDef<Service>[] {
  return [
    {
      accessorKey: "name",
      header: "Service",
      cell: ({ row }) => {
        const s = row.original;
        const thumb = s.images?.[0];
        return (
          <div className="flex items-center gap-3">
            {thumb ? (
              <img
                src={thumb}
                alt={s.name}
                className="h-9 w-9 shrink-0 rounded-md object-cover border"
              />
            ) : (
              <div className="h-9 w-9 shrink-0 rounded-md border bg-muted flex items-center justify-center">
                <ImageOff className="h-4 w-4 text-muted-foreground" />
              </div>
            )}
            <div>
              <p className="font-medium leading-snug">{s.name}</p>
              {s.description && (
                <p className="hidden lg:block text-xs text-muted-foreground line-clamp-1">
                  {s.description}
                </p>
              )}
            </div>
          </div>
        );
      },
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => {
        const cfg = TYPE_CONFIG[row.original.type];
        return (
          <span
            className={`inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium ${cfg.className}`}
          >
            {cfg.label}
          </span>
        );
      },
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const cfg = STATUS_CONFIG[row.original.status];
        return (
          <Badge variant={cfg.variant} className="capitalize">
            {cfg.label}
          </Badge>
        );
      },
    },
    {
      accessorKey: "price",
      header: "Price",
      cell: ({ row }) => {
        const s = row.original;
        return (
          <span className="tabular-nums font-semibold">
            {formatPrice(s.price, s.currency)}
            {s.billing_interval && (
              <span className="ml-1 text-xs font-normal text-muted-foreground">
                /{s.billing_interval}
              </span>
            )}
          </span>
        );
      },
    },
    {
      id: "details",
      header: "Details",
      cell: ({ row }) => {
        const s = row.original;
        if (s.type === "membership") {
          return (
            <div className="flex items-center gap-1 text-sm text-muted-foreground">
              <Users className="h-3.5 w-3.5" />
              {s.billing_interval ?? "—"}
              {s.trial_days != null && (
                <span className="ml-1 text-xs">· {s.trial_days}d trial</span>
              )}
            </div>
          );
        }
        if (s.type === "package") {
          return (
            <div className="flex items-center gap-1 text-sm text-muted-foreground">
              <Clock className="h-3.5 w-3.5" />
              {s.session_count != null ? `${s.session_count} sessions` : "—"}
              {s.validity_days != null && (
                <span className="ml-1 text-xs">· {s.validity_days}d</span>
              )}
            </div>
          );
        }
        return (
          <div className="flex items-center gap-1 text-sm text-muted-foreground">
            <Clock className="h-3.5 w-3.5" />
            {formatDuration(s.duration_min)}
            {s.type === "class" && s.max_concurrent != null && (
              <span className="ml-1 text-xs">· {s.max_concurrent} max</span>
            )}
          </div>
        );
      },
    },
    {
      id: "actions",
      cell: ({ row }) => {
        const s = row.original;
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 cursor-pointer"
                aria-label="Row actions"
              >
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Actions</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="cursor-pointer"
                onClick={() => onEdit(s)}
              >
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem
                className="cursor-pointer"
                onClick={() => onToggleStatus(s)}
              >
                {s.status === "active" ? "Deactivate" : "Activate"}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="cursor-pointer text-destructive focus:text-destructive"
                onClick={() => onDelete(s)}
              >
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
}
