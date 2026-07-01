"use client";

import { useState } from "react";
import {
  IconBrandFacebook,
  IconExternalLink,
  IconLoader2,
  IconUsers,
  IconPlugConnected,
  IconPlugConnectedX,
  IconCopy,
  IconCheck,
  IconAlertTriangle,
} from "@tabler/icons-react";
import { Badge } from "@repo/ui/components/ui/badge";
import { Button } from "@repo/ui/components/ui/button";
import { Skeleton } from "@repo/ui/components/ui/skeleton";
import { getFacebookConnectURL } from "../api/connections";
import {
  useFacebookConnectionStatus,
  useToggleFacebookPage,
} from "../client/connections";
import { ConnectedPage } from "@repo/types";
import { toast } from "@repo/ui/components/ui/sonner";

function CopyableId({ label, value }: { label: string; value: string }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    await navigator.clipboard.writeText(value);
    setCopied(true);
    toast.success(`${label} copied`);
    setTimeout(() => setCopied(false), 1500);
  }

  return (
    <div className="flex items-center justify-between gap-2 rounded-lg bg-muted/50 px-3 py-2 text-sm">
      <div className="min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className="truncate font-mono text-xs">{value}</p>
      </div>
      <button
        type="button"
        onClick={handleCopy}
        className="shrink-0 rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
        aria-label={`Copy ${label}`}
      >
        {copied ? (
          <IconCheck className="size-3.5 text-success" />
        ) : (
          <IconCopy className="size-3.5" />
        )}
      </button>
    </div>
  );
}

function formatDate(value?: string | Date | null) {
  if (!value) return null;
  const date = typeof value === "string" ? new Date(value) : value;
  if (isNaN(date.getTime())) return null;
  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

// Best-effort extraction of a readable message from whatever shape the backend error takes
function extractErrorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  if (err && typeof err === "object") {
    const e = err as Record<string, unknown>;
    if (typeof e.message === "string") return e.message;
    if (typeof e.errors === "string") return e.errors;
    if (e.response && typeof e.response === "object") {
      const r = e.response as Record<string, unknown>;
      if (typeof r.message === "string") return r.message;
    }
  }
  return "Something went wrong. Please try again.";
}

function PageCard({
  page,
  businessId,
}: {
  page: ConnectedPage;
  businessId: string;
}) {
  const toggle = useToggleFacebookPage(businessId);

  const name =
    page.platform_account_name ?? page.platform_account_id ?? "Unknown Page";
  const picture = page.platform_account_image ?? page.details?.picture?.data?.url;
  const followers = page.details?.followers_count ?? page.details?.fan_count;
  const category = page.details?.category;
  const about = page.details?.about;
  const link = page.details?.link;

  const pageId = page.platform_account_id ?? page.id;
  const connectedAt = formatDate(page.connected_at);

  async function handleToggle() {
    const nextIsActive = !page.is_active;
    try {
      const result = await toggle.mutateAsync({ pageId, isActive: nextIsActive });
      const backendMessage =
        (result as { message?: string })?.message ??
        (nextIsActive ? `${name} connected` : `${name} disconnected`);
      toast.success(backendMessage);
    } catch (err) {
      toast.error(extractErrorMessage(err));
    }
  }

  return (
    <div className="flex flex-col gap-5 rounded-2xl border p-6 shadow-sm">
      <div className="flex items-start gap-4">
        {picture ? (
          <img
            src={picture}
            alt={name}
            className="size-16 shrink-0 rounded-full object-cover ring-1 ring-border"
          />
        ) : (
          <div className="flex size-16 shrink-0 items-center justify-center rounded-full bg-primary/10">
            <IconBrandFacebook className="size-7 text-primary" />
          </div>
        )}

        <div className="min-w-0 flex-1 pt-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="truncate text-base font-semibold">{name}</p>
            {page.is_active ? (
              <Badge variant="secondary" className="shrink-0 text-success">
                Connected
              </Badge>
            ) : (
              <Badge variant="outline" className="shrink-0 text-muted-foreground">
                Not connected
              </Badge>
            )}
          </div>
          {category && (
            <p className="mt-0.5 text-sm text-muted-foreground">{category}</p>
          )}
        </div>
      </div>

      {about && (
        <p className="line-clamp-2 text-sm text-muted-foreground">{about}</p>
      )}

      {followers != null && (
        <div className="flex items-center gap-1.5 text-sm">
          <IconUsers className="size-4 text-muted-foreground" />
          <span className="font-medium">{followers.toLocaleString()}</span>
          <span className="text-muted-foreground">followers</span>
        </div>
      )}

      {/* Details block */}
      <div className="space-y-2 border-t pt-4">
        <CopyableId label="Page ID" value={pageId} />

        {page.id && page.id !== pageId && (
          <CopyableId label="Connection ID" value={page.id} />
        )}

        <div className="flex flex-wrap gap-x-6 gap-y-1 px-1 text-xs text-muted-foreground">
          {connectedAt && <span>Connected {connectedAt}</span>}
        </div>
      </div>

      {page.error_message && (
        <div className="flex items-start gap-2 rounded-lg bg-destructive/10 px-3 py-2 text-xs text-destructive">
          <IconAlertTriangle className="mt-0.5 size-3.5 shrink-0" />
          <span>{page.error_message}</span>
        </div>
      )}

      <div className="mt-auto flex items-center gap-2 pt-1">
        <Button
          variant={page.is_active ? "outline" : "default"}
          className={page.is_active ? "flex-1 text-destructive hover:text-destructive" : "flex-1"}
          disabled={toggle.isPending}
          onClick={handleToggle}
        >
          {toggle.isPending ? (
            <IconLoader2 className="size-4 animate-spin" />
          ) : page.is_active ? (
            <IconPlugConnectedX className="size-4" />
          ) : (
            <IconPlugConnected className="size-4" />
          )}
          {toggle.isPending
            ? page.is_active
              ? "Disconnecting…"
              : "Connecting…"
            : page.is_active
              ? "Disconnect"
              : "Connect"}
        </Button>

        {page.is_active && link && (
          <Button asChild variant="outline" size="icon">
            <a href={link} target="_blank" rel="noopener noreferrer" aria-label="View Page">
              <IconExternalLink className="size-4" />
            </a>
          </Button>
        )}
      </div>
    </div>
  );
}

export function FacebookConnection({ businessId }: { businessId: string }) {
  const { data, isLoading } = useFacebookConnectionStatus(businessId);
  const [connecting, setConnecting] = useState(false);

  const status = data?.data;

  async function handleConnect() {
    setConnecting(true);
    try {
      const url = await getFacebookConnectURL(businessId);
      window.location.href = url;
    } catch (err) {
      toast.error(extractErrorMessage(err));
    } finally {
      setConnecting(false);
    }
  }

  if (isLoading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-4 py-24">
        <Skeleton className="size-16 rounded-full" />
        <Skeleton className="h-5 w-40" />
        <Skeleton className="h-4 w-64" />
        <Skeleton className="h-10 w-40 rounded-md" />
      </div>
    );
  }

  if (!status?.connected) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-5 py-24 text-center">
        <div className="flex size-20 items-center justify-center rounded-full bg-primary/10">
          <IconBrandFacebook className="size-10 text-primary" />
        </div>
        <div className="space-y-1">
          <h2 className="text-xl font-semibold">Connect Facebook</h2>
          <p className="max-w-sm text-sm text-muted-foreground">
            Link your Facebook account to manage pages, publish content, and
            track engagement from one place.
          </p>
        </div>
        <Button size="lg" onClick={handleConnect} disabled={connecting}>
          {connecting ? (
            <IconLoader2 className="size-4 animate-spin" />
          ) : (
            <IconBrandFacebook className="size-4" />
          )}
          {connecting ? "Redirecting…" : "Connect Now"}
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <div className="flex size-10 items-center justify-center rounded-full bg-primary/10">
          <IconBrandFacebook className="size-5 text-primary" />
        </div>
        <div>
          <h2 className="font-semibold">Facebook Pages</h2>
          <p className="text-sm text-muted-foreground">
            {status.pages.length} page{status.pages.length !== 1 ? "s" : ""} found
          </p>
        </div>
      </div>

      <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {status.pages.map((page) => (
          <PageCard key={page.id} page={page} businessId={businessId} />
        ))}
      </div>
    </div>
  );
}