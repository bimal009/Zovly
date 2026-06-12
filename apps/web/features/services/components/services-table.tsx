"use client";

import * as React from "react";
import { Calendar, DollarSign, Plus, TrendingUp, Users } from "lucide-react";
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
  useCreateService,
  useDeleteService,
  useGetServices,
  useUpdateService,
} from "../client/services";
import type {
  CreateServiceInput,
  Service,
  ServiceStatus,
  ServiceType,
  UpdateServiceInput,
} from "../api/services";
import { TYPE_CONFIG, getServiceColumns } from "./columns";
import { ServiceFormDialog } from "./service-form-dialog";

const TYPE_VALUES = ["appointment", "class", "membership", "package"] as const;
const STATUS_VALUES = ["active", "inactive", "archived"] as const;

export function ServicesTable() {
  const [filters, setFilters] = useQueryStates({
    search: parseAsString.withDefault(""),
    type: parseAsStringEnum<ServiceType>([...TYPE_VALUES]),
    status: parseAsStringEnum<ServiceStatus>([...STATUS_VALUES]),
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
    ...(filters.type ? { type: filters.type } : {}),
    ...(filters.status ? { status: filters.status } : {}),
    ...(filters.search ? { search: filters.search } : {}),
    limit: filters.limit,
    offset,
  };

  const { data, isLoading } = useGetServices(params);
  const createMutation = useCreateService();
  const updateMutation = useUpdateService();
  const deleteMutation = useDeleteService();

  const services = data?.data ?? [];
  const total = data?.meta?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / filters.limit));

  const [formOpen, setFormOpen] = React.useState(false);
  const [editing, setEditing] = React.useState<Service | null>(null);
  const [deleteTarget, setDeleteTarget] = React.useState<Service | null>(null);

  const saving = createMutation.isPending || updateMutation.isPending;
  const deleting = deleteMutation.isPending;

  // ── Stats ──────────────────────────────────────────────────────────────────
  const activeServices = services.filter((s) => s.status === "active");
  const activeCount = activeServices.length;
  const avgPrice =
    activeCount > 0
      ? activeServices.reduce((sum, s) => sum + s.price, 0) / activeCount
      : 0;

  const typeCounts = (Object.keys(TYPE_CONFIG) as ServiceType[]).reduce(
    (acc, t) => {
      acc[t] = activeServices.filter((s) => s.type === t).length;
      return acc;
    },
    {} as Record<ServiceType, number>,
  );

  const typeBreakdown =
    (Object.entries(typeCounts) as [ServiceType, number][])
      .filter(([, c]) => c > 0)
      .map(([t, c]) => `${c} ${TYPE_CONFIG[t].label}`)
      .join(" · ") || "None active";

  const stats = [
    {
      label: "Total Services",
      value: isLoading ? "—" : total,
      icon: Calendar,
    },
    { label: "Active", value: isLoading ? "—" : activeCount, icon: TrendingUp },
    {
      label: "Avg Price",
      value: isLoading
        ? "—"
        : new Intl.NumberFormat("en-US", {
            style: "currency",
            currency: "USD",
            maximumFractionDigits: 0,
          }).format(avgPrice / 100),
      icon: DollarSign,
    },
    {
      label: "By Type",
      value: isLoading ? "—" : activeCount,
      sub: isLoading ? undefined : typeBreakdown,
      icon: Users,
    },
  ];

  // ── Handlers ───────────────────────────────────────────────────────────────
  function openCreate() {
    setEditing(null);
    setFormOpen(true);
  }

  function openEdit(s: Service) {
    setEditing(s);
    setFormOpen(true);
  }

  function handleSave(
    data: CreateServiceInput | { id: string; input: UpdateServiceInput },
  ) {
    if ("id" in data) {
      updateMutation.mutate(data, { onSuccess: () => setFormOpen(false) });
    } else {
      createMutation.mutate(data, { onSuccess: () => setFormOpen(false) });
    }
  }

  function handleToggleStatus(s: Service) {
    updateMutation.mutate({
      id: s.id,
      input: { status: s.status === "active" ? "inactive" : "active" },
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
      getServiceColumns({
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
        value={filters.type ?? "all"}
        onValueChange={(v) =>
          setFilters({
            type: v === "all" ? null : (v as ServiceType),
            page: 1,
          })
        }
      >
        <SelectTrigger className="w-36">
          <SelectValue placeholder="Type" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Types</SelectItem>
          <SelectItem value="appointment">Appointment</SelectItem>
          <SelectItem value="class">Class</SelectItem>
          <SelectItem value="membership">Membership</SelectItem>
          <SelectItem value="package">Package</SelectItem>
        </SelectContent>
      </Select>
      <Select
        value={filters.status ?? "all"}
        onValueChange={(v) =>
          setFilters({
            status: v === "all" ? null : (v as ServiceStatus),
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
        Add Service
      </Button>
    </div>
  );

  return (
    <>
      <SectionCards cards={stats} cols={4} />

      <DataTable
        data={services}
        columns={columns}
        searchPlaceholder="Search services…"
        searchValue={searchInput}
        onSearchChange={handleSearchChange}
        pageIndex={filters.page - 1}
        pageSize={filters.limit}
        pageCount={pageCount}
        onPageChange={(idx) => setFilters({ page: idx + 1 })}
        onPageSizeChange={(size) => setFilters({ limit: size, page: 1 })}
        loading={isLoading}
        actions={tableActions}
        emptyMessage="No services found."
      />

      <ServiceFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        editing={editing}
        onSave={handleSave}
        saving={saving}
      />

      <ConfirmDeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete service?"
        description={
          <>
            This will permanently delete{" "}
            <span className="font-semibold">{deleteTarget?.name}</span>. All
            associated booking data may be affected.
          </>
        }
        onConfirm={handleConfirmDelete}
        loading={deleting}
      />
    </>
  );
}
