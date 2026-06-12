"use client";

import * as React from "react";
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnDef,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
} from "@tanstack/react-table";

// Re-export so consumers don't need @tanstack/react-table as a direct dep
export type { ColumnDef } from "@tanstack/react-table";
import {
  IconChevronDown,
  IconChevronLeft,
  IconChevronRight,
  IconChevronsLeft,
  IconChevronsRight,
  IconLayoutColumns,
  IconSearch,
  IconX,
} from "@tabler/icons-react";

import { Button } from "./button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "./dropdown-menu";
import { Input } from "./input";
import { Label } from "./label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./select";
import { Skeleton } from "./skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "./table";

export type DataTableProps<TData> = {
  data: TData[];
  columns: ColumnDef<TData>[];
  searchKey?: string;
  searchPlaceholder?: string;
  /** Controlled search value — bypasses TanStack column filter and calls onSearchChange instead */
  searchValue?: string;
  onSearchChange?: (value: string) => void;
  /** Controlled pagination — when provided the table is in server-side pagination mode */
  pageIndex?: number;
  pageSize?: number;
  pageCount?: number;
  onPageChange?: (pageIndex: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  loading?: boolean;
  loadingRows?: number;
  actions?: React.ReactNode;
  emptyMessage?: string;
  showColumnToggle?: boolean;
};

export function DataTable<TData>({
  data,
  columns,
  searchKey,
  searchPlaceholder = "Search…",
  searchValue: externalSearchValue,
  onSearchChange,
  pageIndex: externalPageIndex,
  pageSize: externalPageSize,
  pageCount: externalPageCount,
  onPageChange,
  onPageSizeChange,
  loading = false,
  loadingRows = 5,
  actions,
  emptyMessage = "No results.",
  showColumnToggle = true,
}: DataTableProps<TData>) {
  const isControlledPagination = onPageChange !== undefined;

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] =
    React.useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});
  const [internalPagination, setInternalPagination] = React.useState({
    pageIndex: 0,
    pageSize: 10,
  });

  const pagination = isControlledPagination
    ? { pageIndex: externalPageIndex ?? 0, pageSize: externalPageSize ?? 10 }
    : internalPagination;

  const table = useReactTable({
    data,
    columns,
    pageCount: isControlledPagination ? (externalPageCount ?? -1) : undefined,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
      pagination,
    },
    manualPagination: isControlledPagination,
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onPaginationChange: (updater) => {
      const next =
        typeof updater === "function" ? updater(pagination) : updater;
      if (isControlledPagination) {
        if (next.pageIndex !== pagination.pageIndex) onPageChange(next.pageIndex);
        if (next.pageSize !== pagination.pageSize) onPageSizeChange?.(next.pageSize);
      } else {
        setInternalPagination(next);
      }
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  const isControlled = onSearchChange !== undefined;
  const internalSearchValue = searchKey
    ? ((table.getColumn(searchKey)?.getFilterValue() as string) ?? "")
    : "";
  const searchValue = isControlled
    ? (externalSearchValue ?? "")
    : internalSearchValue;

  const hasToolbar = searchKey || onSearchChange || actions || showColumnToggle;

  return (
    <div className="flex flex-col gap-4">
      {/* Toolbar */}
      {hasToolbar && (
        <div className="flex flex-wrap items-center justify-between gap-3 px-4 lg:px-6">
          {(searchKey || onSearchChange) && (
            <div className="relative flex-1 min-w-[160px] max-w-sm">
              <IconSearch className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder={searchPlaceholder}
                className="pl-9 pr-8"
                value={searchValue}
                onChange={(e) => {
                  if (isControlled) {
                    onSearchChange!(e.target.value);
                  } else if (searchKey) {
                    table.getColumn(searchKey)?.setFilterValue(e.target.value);
                  }
                }}
              />
              {searchValue && (
                <button
                  onClick={() => {
                    if (isControlled) {
                      onSearchChange!("");
                    } else if (searchKey) {
                      table.getColumn(searchKey)?.setFilterValue("");
                    }
                  }}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                  aria-label="Clear search"
                >
                  <IconX className="size-4" />
                </button>
              )}
            </div>
          )}
          <div className="flex items-center gap-2 ml-auto">
            {showColumnToggle && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="outline"
                    size="sm"
                    className="cursor-pointer"
                  >
                    <IconLayoutColumns className="size-4" />
                    <span className="hidden lg:inline">Columns</span>
                    <IconChevronDown className="size-3" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-48">
                  {table
                    .getAllColumns()
                    .filter(
                      (col) =>
                        typeof col.accessorFn !== "undefined" &&
                        col.getCanHide()
                    )
                    .map((col) => (
                      <DropdownMenuCheckboxItem
                        key={col.id}
                        className="capitalize"
                        checked={col.getIsVisible()}
                        onCheckedChange={(value) =>
                          col.toggleVisibility(!!value)
                        }
                      >
                        {col.id}
                      </DropdownMenuCheckboxItem>
                    ))}
                </DropdownMenuContent>
              </DropdownMenu>
            )}
            {actions}
          </div>
        </div>
      )}

      {/* Table */}
      <div className="overflow-hidden rounded-lg border mx-4 lg:mx-6">
        <Table>
          <TableHeader className="bg-muted/40">
            {table.getHeaderGroups().map((hg) => (
              <TableRow key={hg.id}>
                {hg.headers.map((header) => (
                  <TableHead key={header.id} colSpan={header.colSpan}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              Array.from({ length: loadingRows }).map((_, i) => (
                <TableRow key={i}>
                  {columns.map((_, j) => (
                    <TableCell key={j}>
                      <Skeleton className="h-4 w-full" />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : table.getRowModel().rows.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center text-muted-foreground"
                >
                  {emptyMessage}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between px-4 pb-2 lg:px-6">
        <div className="hidden text-sm text-muted-foreground lg:block">
          {table.getFilteredSelectedRowModel().rows.length > 0 ? (
            <span>
              {table.getFilteredSelectedRowModel().rows.length} of{" "}
              {table.getFilteredRowModel().rows.length} row(s) selected
            </span>
          ) : (
            <span>
              {table.getFilteredRowModel().rows.length} row(s)
            </span>
          )}
        </div>
        <div className="flex w-full items-center gap-6 lg:w-fit">
          <div className="hidden items-center gap-2 lg:flex">
            <Label
              htmlFor="rows-per-page"
              className="text-sm font-medium whitespace-nowrap"
            >
              Rows per page
            </Label>
            <Select
              value={`${table.getState().pagination.pageSize}`}
              onValueChange={(v) => table.setPageSize(Number(v))}
            >
              <SelectTrigger size="sm" className="w-20" id="rows-per-page">
                <SelectValue />
              </SelectTrigger>
              <SelectContent side="top">
                {[10, 20, 30, 50].map((n) => (
                  <SelectItem key={n} value={`${n}`}>
                    {n}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex w-fit items-center justify-center text-sm font-medium">
            Page {table.getState().pagination.pageIndex + 1} of{" "}
            {table.getPageCount() || 1}
          </div>
          <div className="ml-auto flex items-center gap-2 lg:ml-0">
            <Button
              variant="outline"
              className="hidden size-8 lg:flex cursor-pointer"
              size="icon"
              onClick={() => table.setPageIndex(0)}
              disabled={!table.getCanPreviousPage()}
              aria-label="First page"
            >
              <IconChevronsLeft className="size-4" />
            </Button>
            <Button
              variant="outline"
              className="size-8 cursor-pointer"
              size="icon"
              onClick={() => table.previousPage()}
              disabled={!table.getCanPreviousPage()}
              aria-label="Previous page"
            >
              <IconChevronLeft className="size-4" />
            </Button>
            <Button
              variant="outline"
              className="size-8 cursor-pointer"
              size="icon"
              onClick={() => table.nextPage()}
              disabled={!table.getCanNextPage()}
              aria-label="Next page"
            >
              <IconChevronRight className="size-4" />
            </Button>
            <Button
              variant="outline"
              className="hidden size-8 lg:flex cursor-pointer"
              size="icon"
              onClick={() => table.setPageIndex(table.getPageCount() - 1)}
              disabled={!table.getCanNextPage()}
              aria-label="Last page"
            >
              <IconChevronsRight className="size-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
