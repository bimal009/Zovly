"use client";

import { useState } from "react";
import {
  IconBrandFacebook,
  IconBrandInstagram,
  IconExternalLink,
  IconInfoCircle,
  IconLoader2,
  IconUsers,
} from "@tabler/icons-react";
import { Badge } from "@repo/ui/components/ui/badge";
import { Button } from "@repo/ui/components/ui/button";
import { Skeleton } from "@repo/ui/components/ui/skeleton";
import { ConfirmDeleteDialog } from "@/components/confirm-delete-dialog";
import { getFacebookConnectURL } from "../api/connections";
import {
  useFacebookConnectionStatus,
  useToggleFacebookPage,
} from "../client/connections";
import type { ConnectedPage } from "../types/connections";

function PageCard({ page }: { page: ConnectedPage }) {
  const [confirmOpen, setConfirmOpen] = useState(false);
  const toggle = useToggleFacebookPage();

  const name =
    page.platform_account_name ?? page.platform_account_id ?? "Unknown Page";
  const picture = page.details?.picture?.data?.url;
  const followers = page.details?.followers_count ?? page.details?.fan_count;
  const category = page.details?.category;
  const about = page.details?.about;
  const link = page.details?.link;

  return (
    <>
      <div className="flex flex-col gap-4 rounded-xl border p-5">
        <div className="flex items-start gap-4">
          {picture ? (
            <img
              src={picture}
              alt={name}
              className="size-12 shrink-0 rounded-full object-cover"
            />
          ) : (
            <div className="flex size-12 shrink-0 items-center justify-center rounded-full bg-[#1877F2]/10">
              <IconBrandFacebook className="size-6 text-[#1877F2]" />
            </div>
          )}

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="truncate font-semibold">{name}</p>
              {page.is_active ? (
                <Badge variant="secondary" className="shrink-0 text-green-600">
                  Active
                </Badge>
              ) : (
                <Badge
                  variant="outline"
                  className="shrink-0 text-muted-foreground"
                >
                  Inactive
                </Badge>
              )}
            </div>
            {category && (
              <p className="text-muted-foreground text-xs">{category}</p>
            )}
          </div>
        </div>

        {about && (
          <p className="text-muted-foreground line-clamp-2 text-sm">{about}</p>
        )}

        {followers != null && (
          <div className="flex items-center gap-1.5 text-sm">
            <IconUsers className="text-muted-foreground size-4" />
            <span className="font-medium">{followers.toLocaleString()}</span>
            <span className="text-muted-foreground">followers</span>
          </div>
        )}

        <div className="flex gap-2">
          {page.is_active ? (
            <Button asChild variant="outline" size="sm" className="flex-1">
              <a href={link ?? "#"} target="_blank" rel="noopener noreferrer">
                <IconExternalLink className="size-3.5" />
                View Details
              </a>
            </Button>
          ) : (
            <Button
              size="sm"
              className="flex-1"
              onClick={() => setConfirmOpen(true)}
              disabled={toggle.isPending}
            >
              {toggle.isPending ? (
                <IconLoader2 className="size-3.5 animate-spin" />
              ) : (
                <IconBrandFacebook className="size-3.5" />
              )}
              Connect
            </Button>
          )}
        </div>
      </div>

      <ConfirmDeleteDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title="Connect this page?"
        description={`Are you sure you want to connect "${name}" to your workspace?`}
        confirmText="Connect"
        onConfirm={() => {
          toggle.mutate(page.platform_account_id ?? page.id, {
            onSuccess: () => setConfirmOpen(false),
          });
        }}
        loading={toggle.isPending}
      />
    </>
  );
}

export function FacebookConnection() {
  const { data, isLoading } = useFacebookConnectionStatus();
  const [connecting, setConnecting] = useState(false);

  const status = data?.data;

  async function handleConnect() {
    setConnecting(true);
    try {
      const url = await getFacebookConnectURL();
      window.location.href = url;
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
        <div className="flex size-20 items-center justify-center rounded-full bg-[#1877F2]/10">
          <IconBrandFacebook className="size-10 text-[#1877F2]" />
        </div>
        <div className="space-y-1">
          <h2 className="text-xl font-semibold">Connect Facebook</h2>
          <p className="text-muted-foreground max-w-sm text-sm">
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
        <div className="flex size-10 items-center justify-center rounded-full bg-[#1877F2]/10">
          <IconBrandFacebook className="size-5 text-[#1877F2]" />
        </div>
        <div>
          <h2 className="font-semibold">Facebook Pages</h2>
          <p className="text-muted-foreground text-sm">
            {status.pages.length} page{status.pages.length !== 1 ? "s" : ""}{" "}
            found
          </p>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {status.pages.map((page) => (
          <PageCard key={page.id} page={page} />
        ))}
      </div>
    </div>
  );
}
