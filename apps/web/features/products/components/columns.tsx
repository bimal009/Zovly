"use client";

import { AlertTriangle, MoreHorizontal, Package } from "lucide-react";
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
import type { Product, ProductStatus } from "../types/products";

const STATUS_CONFIG: Record<
  ProductStatus,
  { label: string; variant: "default" | "secondary" | "outline" }
> = {
  active: { label: "Active", variant: "default" },
  inactive: { label: "Inactive", variant: "secondary" },
  archived: { label: "Archived", variant: "outline" },
};

function formatPrice(cents: number, currency: string) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    minimumFractionDigits: 2,
  }).format(cents / 100);
}

interface ColumnActions {
  onEdit: (p: Product) => void;
  onDelete: (p: Product) => void;
  onToggleStatus: (p: Product) => void;
}

export function getProductColumns({
  onEdit,
  onDelete,
  onToggleStatus,
}: ColumnActions): ColumnDef<Product>[] {
  return [
    {
      accessorKey: "name",
      header: "Product",
      cell: ({ row }) => {
        const p = row.original;
        const thumb = p.images?.[0];
        return (
          <div className="flex items-center gap-3">
            {thumb ? (
              <img
                src={thumb}
                alt={p.name}
                className="h-9 w-9 shrink-0 rounded-md object-cover border"
              />
            ) : (
              <div className="h-9 w-9 shrink-0 rounded-md border bg-muted flex items-center justify-center">
                <Package className="h-4 w-4 text-muted-foreground" />
              </div>
            )}
            <div>
              <p className="font-medium leading-snug">{p.name}</p>
              {p.description && (
                <p className="hidden lg:block text-xs text-muted-foreground line-clamp-1">
                  {p.description}
                </p>
              )}
            </div>
          </div>
        );
      },
    },
    {
      accessorKey: "sku",
      header: "SKU",
      cell: ({ row }) => (
        <span className="font-mono text-xs text-muted-foreground">
          {row.original.sku ?? "—"}
        </span>
      ),
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
        const p = row.original;
        return (
          <div className="flex items-center gap-1.5">
            <span className="tabular-nums font-semibold">
              {formatPrice(p.price, p.currency)}
            </span>
            {p.discount > 0 && (
              <Badge variant="secondary" className="text-xs px-1.5 py-0">
                {p.discount}% off
              </Badge>
            )}
          </div>
        );
      },
    },
    {
      accessorKey: "stock_qty",
      header: "Stock",
      cell: ({ row }) => {
        const p = row.original;
        const isLow =
          p.low_stock_threshold != null &&
          p.stock_qty <= p.low_stock_threshold;
        return (
          <div className="flex items-center gap-1.5">
            {isLow && (
              <AlertTriangle className="text-warning h-3.5 w-3.5 shrink-0" />
            )}
            <span
              className={`text-sm tabular-nums ${isLow ? "text-warning font-semibold" : ""}`}
            >
              {p.stock_qty} units
            </span>
          </div>
        );
      },
    },
    {
      id: "actions",
      cell: ({ row }) => {
        const p = row.original;
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
                onClick={() => onEdit(p)}
              >
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem
                className="cursor-pointer"
                onClick={() => onToggleStatus(p)}
              >
                {p.status === "active" ? "Deactivate" : "Activate"}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="cursor-pointer text-destructive focus:text-destructive"
                onClick={() => onDelete(p)}
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
