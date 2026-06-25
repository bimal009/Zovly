"use client";

import * as React from "react";
import { Plus, Tag } from "lucide-react";
import { Badge } from "@repo/ui/components/ui/badge";
import { Button } from "@repo/ui/components/ui/button";
import { DataTable, type ColumnDef } from "@repo/ui/components/ui/data-table";
import { useCreateCategory, useGetCategories } from "../client/categories";
import type { Category, CreateCategoryInput } from "../types/categories";
import { CategoryFormDialog } from "./category-form-dialog";

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

const columns: ColumnDef<Category>[] = [
  {
    accessorKey: "name",
    header: "Category",
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border bg-muted">
          <Tag className="h-4 w-4 text-muted-foreground" />
        </div>
        <div>
          <p className="font-medium capitalize leading-snug">
            {row.original.name.toLowerCase()}
          </p>
          {row.original.description && (
            <p className="hidden text-xs text-muted-foreground line-clamp-1 lg:block">
              {row.original.description}
            </p>
          )}
        </div>
      </div>
    ),
  },
  {
    accessorKey: "slug",
    header: "Slug",
    cell: ({ row }) =>
      row.original.slug ? (
        <Badge variant="outline" className="font-mono text-xs">
          {row.original.slug}
        </Badge>
      ) : (
        <span className="text-muted-foreground">—</span>
      ),
  },
  {
    accessorKey: "created_at",
    header: "Created",
    cell: ({ row }) => (
      <span className="text-sm text-muted-foreground tabular-nums">
        {formatDate(row.original.created_at)}
      </span>
    ),
  },
];

export function CategoriesTable() {
  const { data, isLoading } = useGetCategories();
  const createMutation = useCreateCategory();

  const categories = data?.data ?? [];

  const [formOpen, setFormOpen] = React.useState(false);

  function handleSave(input: CreateCategoryInput) {
    createMutation.mutate(input, { onSuccess: () => setFormOpen(false) });
  }

  const tableActions = (
    <Button onClick={() => setFormOpen(true)} className="cursor-pointer">
      <Plus className="mr-2 h-4 w-4" />
      Add Category
    </Button>
  );

  return (
    <>
      <DataTable
        data={categories}
        columns={columns}
        searchKey="name"
        searchPlaceholder="Search categories…"
        loading={isLoading}
        actions={tableActions}
        emptyMessage="No categories yet. Create one to organize your products."
      />

      <CategoryFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        onSave={handleSave}
        saving={createMutation.isPending}
      />
    </>
  );
}
