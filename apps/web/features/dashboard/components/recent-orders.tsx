"use client";

import { Badge } from "@repo/ui/components/ui/badge";
import { DataTable, type ColumnDef } from "@repo/ui/components/ui/data-table";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@repo/ui/components/ui/card";

// ─── Types ────────────────────────────────────────────────────────────────────

type LeadStatus = "new" | "qualified" | "converted" | "lost";
type Platform = "Instagram" | "WhatsApp" | "Facebook" | "TikTok";

type Lead = {
  id: string;
  name: string;
  platform: Platform;
  message: string;
  score: number;
  status: LeadStatus;
  date: string;
};

// ─── Data ─────────────────────────────────────────────────────────────────────

const leads: Lead[] = [
  {
    id: "L-001",
    name: "Sarah Chen",
    platform: "Instagram",
    message: "Hi! Do you still have the blue jacket in size M?",
    score: 92,
    status: "converted",
    date: "Jun 12, 2026",
  },
  {
    id: "L-002",
    name: "James Wilson",
    platform: "WhatsApp",
    message: "Can I book a consultation for next Tuesday?",
    score: 88,
    status: "qualified",
    date: "Jun 11, 2026",
  },
  {
    id: "L-003",
    name: "Maria Garcia",
    platform: "Instagram",
    message: "What are your prices for the premium plan?",
    score: 75,
    status: "new",
    date: "Jun 11, 2026",
  },
  {
    id: "L-004",
    name: "David Kim",
    platform: "Facebook",
    message: "Is this service available in my area?",
    score: 61,
    status: "qualified",
    date: "Jun 10, 2026",
  },
  {
    id: "L-005",
    name: "Emma Davis",
    platform: "TikTok",
    message: "Saw your video! How do I get started?",
    score: 84,
    status: "new",
    date: "Jun 10, 2026",
  },
  {
    id: "L-006",
    name: "Carlos Rodriguez",
    platform: "WhatsApp",
    message: "Need 50 units for my store. What's wholesale pricing?",
    score: 95,
    status: "converted",
    date: "Jun 9, 2026",
  },
  {
    id: "L-007",
    name: "Priya Singh",
    platform: "Instagram",
    message: "Do you ship internationally?",
    score: 42,
    status: "lost",
    date: "Jun 9, 2026",
  },
  {
    id: "L-008",
    name: "Alex Thompson",
    platform: "Facebook",
    message: "What's included in the starter package?",
    score: 79,
    status: "qualified",
    date: "Jun 8, 2026",
  },
];

// ─── Config ───────────────────────────────────────────────────────────────────

const statusVariant: Record<
  LeadStatus,
  "default" | "secondary" | "destructive" | "outline"
> = {
  converted: "default",
  qualified: "secondary",
  new: "outline",
  lost: "destructive",
};

const platformColor: Record<Platform, string> = {
  Instagram: "text-pink-600",
  WhatsApp: "text-green-600",
  Facebook: "text-blue-600",
  TikTok: "text-foreground",
};

// ─── Columns ──────────────────────────────────────────────────────────────────

const columns: ColumnDef<Lead>[] = [
  {
    accessorKey: "name",
    header: "Contact",
    cell: ({ row }) => (
      <span className="font-medium">{row.original.name}</span>
    ),
  },
  {
    accessorKey: "platform",
    header: "Platform",
    cell: ({ row }) => (
      <span
        className={`text-sm font-medium ${platformColor[row.original.platform]}`}
      >
        {row.original.platform}
      </span>
    ),
  },
  {
    accessorKey: "message",
    header: "Message",
    cell: ({ row }) => (
      <span className="hidden max-w-[280px] truncate text-muted-foreground lg:block">
        {row.original.message}
      </span>
    ),
  },
  {
    accessorKey: "score",
    header: "Score",
    cell: ({ row }) => {
      const score = row.original.score;
      const color =
        score >= 80
          ? "text-green-600"
          : score >= 60
            ? "text-yellow-600"
            : "text-muted-foreground";
      return (
        <span className={`text-sm font-semibold tabular-nums ${color}`}>
          {score}
        </span>
      );
    },
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => (
      <Badge
        variant={statusVariant[row.original.status]}
        className="capitalize"
      >
        {row.original.status}
      </Badge>
    ),
  },
  {
    accessorKey: "date",
    header: "Date",
    cell: ({ row }) => (
      <span className="text-muted-foreground">{row.original.date}</span>
    ),
  },
];

// ─── Component ────────────────────────────────────────────────────────────────

export function RecentOrders() {
  return (
    <div className="px-4 lg:px-6">
      <Card>
        <CardHeader>
          <CardTitle>Recent Leads</CardTitle>
          <CardDescription>
            AI-captured leads from DMs and comments across platforms
          </CardDescription>
        </CardHeader>
        <CardContent className="p-0 pb-4">
          <DataTable
            data={leads}
            columns={columns}
            searchKey="name"
            searchPlaceholder="Search leads…"
            showColumnToggle={false}
            emptyMessage="No leads yet."
          />
        </CardContent>
      </Card>
    </div>
  );
}
