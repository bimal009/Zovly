"use client";

import * as React from "react";
import {
  AlertTriangle,
  DollarSign,
  Package,
  Plus,
  TrendingUp,
} from "lucide-react";
import {
  useQueryStates,
  parseAsString,
  parseAsStringEnum,
  parseAsInteger,
} from "nuqs";
import { useDebouncedCallback } from "use-debounce";
import { Button } from "@repo/ui/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@repo/ui/components/ui/select";
import { DataTable } from "@repo/ui/components/ui/data-table";
import { SectionCards } from "@repo/ui/components/ui/section-cards";
import { ConfirmDeleteDialog } from "@/components/confirm-delete-dialog";
import {
  useCreateProduct,
  useDeleteProduct,
  useGetProducts,
  useUpdateProduct,
} from "../client/products";
import type {
  Product,
  ProductStatus,
  UpdateProductInput,
  CreateProductInput,
} from "../types/products";
import { getProductColumns } from "./columns";
import { ProductFormDialog } from "./product-form-dialog";

const STATUS_VALUES = ["active", "inactive", "archived"] as const;
const PAGE_SIZE_VALUES = [10, 20, 50, 100] as const;

export function ProductsTable() {
  const [filters, setFilters] = useQueryStates({
    search: parseAsString.withDefault(""),
    status: parseAsStringEnum<ProductStatus>([...STATUS_VALUES]),
    page: parseAsInteger.withDefault(1),
    limit: parseAsInteger.withDefault(10),
  });

  const [searchInput, setSearchInput] = React.useState(filters.search);

  const flushSearch = useDebouncedCallback((val: string) => {
    setFilters({ search: val || null, page: 1 });
  }, 300);

  function handleSearchChange(val: string) {
    setSearchInput(val);
    flushSearch(val);
  }

  const offset = (filters.page - 1) * filters.limit;

  const params = {
    ...(filters.status ? { status: filters.status } : {}),
    ...(filters.search ? { search: filters.search } : {}),
    limit: filters.limit,
    offset,
  };

  const { data, isLoading } = useGetProducts(params);
  const createMutation = useCreateProduct();
  const updateMutation = useUpdateProduct();
  const deleteMutation = useDeleteProduct();

  const products = data?.data ?? [];
  const total = data?.meta?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / filters.limit));

  const [formOpen, setFormOpen] = React.useState(false);
  const [editing, setEditing] = React.useState<Product | null>(null);
  const [deleteTarget, setDeleteTarget] = React.useState<Product | null>(null);

  const saving = createMutation.isPending || updateMutation.isPending;
  const deleting = deleteMutation.isPending;

  // ── Stats ──────────────────────────────────────────────────────────────────
  const activeCount = products.filter((p) => p.status === "active").length;
  const lowStockCount = products.filter(
    (p) =>
      p.low_stock_threshold != null && p.stock_qty <= p.low_stock_threshold,
  ).length;
  const catalogValue = products
    .filter((p) => p.status === "active")
    .reduce((sum, p) => sum + p.price * p.stock_qty, 0);

  const stats = [
    { label: "Total Products", value: isLoading ? "—" : total, icon: Package },
    { label: "Active", value: isLoading ? "—" : activeCount, icon: TrendingUp },
    {
      label: "Low Stock",
      value: isLoading ? "—" : lowStockCount,
      icon: AlertTriangle,
      description: lowStockCount > 0 ? `${lowStockCount} need restocking` : undefined,
      trendUp: lowStockCount === 0,
    },
    {
      label: "Catalog Value",
      value: isLoading
        ? "—"
        : new Intl.NumberFormat("en-US", {
            style: "currency",
            currency: "USD",
            maximumFractionDigits: 0,
          }).format(catalogValue / 100),
      icon: DollarSign,
    },
  ];

  // ── Handlers ───────────────────────────────────────────────────────────────
  function openCreate() {
    setEditing(null);
    setFormOpen(true);
  }

  function openEdit(p: Product) {
    setEditing(p);
    setFormOpen(true);
  }

  function handleSave(
    data: CreateProductInput | { id: string; input: UpdateProductInput },
  ) {
    if ("id" in data) {
      updateMutation.mutate(data, { onSuccess: () => setFormOpen(false) });
    } else {
      createMutation.mutate(data, { onSuccess: () => setFormOpen(false) });
    }
  }

  function handleToggleStatus(p: Product) {
    updateMutation.mutate({
      id: p.id,
      input: { status: p.status === "active" ? "inactive" : "active" },
    });
  }

  function handleConfirmDelete() {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id, {
      onSuccess: () => setDeleteTarget(null),
    });
  }

  const columns = React.useMemo(
    () =>
      getProductColumns({
        onEdit: openEdit,
        onDelete: setDeleteTarget,
        onToggleStatus: handleToggleStatus,
      }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  const tableActions = (
    <div className="flex items-center gap-2">
      <Select
        value={filters.status ?? "all"}
        onValueChange={(v) =>
          setFilters({
            status: v === "all" ? null : (v as ProductStatus),
            page: 1,
          })
        }
      >
        <SelectTrigger className="w-36">
          <SelectValue placeholder="Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Status</SelectItem>
          <SelectItem value="active">Active</SelectItem>
          <SelectItem value="inactive">Inactive</SelectItem>
          <SelectItem value="archived">Archived</SelectItem>
        </SelectContent>
      </Select>
      <Button onClick={openCreate} className="cursor-pointer">
        <Plus className="mr-2 h-4 w-4" />
        Add Product
      </Button>
    </div>
  );

  return (
    <>
      <SectionCards cards={stats} cols={4} />

      <DataTable
        data={products}
        columns={columns}
        searchPlaceholder="Search name or SKU…"
        searchValue={searchInput}
        onSearchChange={handleSearchChange}
        pageIndex={filters.page - 1}
        pageSize={filters.limit}
        pageCount={pageCount}
        onPageChange={(idx) => setFilters({ page: idx + 1 })}
        onPageSizeChange={(size) => setFilters({ limit: size, page: 1 })}
        loading={isLoading}
        actions={tableActions}
        emptyMessage="No products found."
      />

      <ProductFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        editing={editing}
        onSave={handleSave}
        saving={saving}
      />

      <ConfirmDeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete product?"
        description={
          <>
            This will permanently delete{" "}
            <span className="font-semibold">{deleteTarget?.name}</span>. This
            action cannot be undone.
          </>
        }
        onConfirm={handleConfirmDelete}
        loading={deleting}
      />
    </>
  );
}
