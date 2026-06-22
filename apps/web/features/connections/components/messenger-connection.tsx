"use client";

import { useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import {
  IconBrandFacebook,
  IconBrandMessenger,
  IconCheck,
  IconExternalLink,
  IconLoader2,
  IconUsers,
} from "@tabler/icons-react";
import { Badge } from "@repo/ui/components/ui/badge";
import { Button } from "@repo/ui/components/ui/button";
import { Skeleton } from "@repo/ui/components/ui/skeleton";
import { ConfirmDeleteDialog } from "@/components/confirm-delete-dialog";
import {
  useMessengerConnectionStatus,
  useSubscribeMessengerPage,
} from "../client/connections";
import type { ConnectedPage } from "../types/connections";

function MessengerPageCard({ page }: { page: ConnectedPage }) {
  const [confirmOpen, setConfirmOpen] = useState(false);
  const subscribe = useSubscribeMessengerPage();

  const name =
    page.platform_account_name ?? page.platform_account_id ?? "Unknown Page";
  const picture = page.details?.picture?.data?.url;
  const followers = page.details?.followers_count ?? page.details?.fan_count;
  const category = page.details?.category;
  const isSubscribed = !!page.webhook_subscribed_at;

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
            <div className="flex size-12 shrink-0 items-center justify-center rounded-full bg-primary/10">
              <IconBrandMessenger className="size-6 text-primary" />
            </div>
          )}

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="truncate font-semibold">{name}</p>
              {isSubscribed ? (
                <Badge variant="secondary" className="text-success shrink-0">
                  Connected
                </Badge>
              ) : (
                <Badge
                  variant="outline"
                  className="shrink-0 text-muted-foreground"
                >
                  Not Connected
                </Badge>
              )}
            </div>
            {category && (
              <p className="text-muted-foreground text-xs">{category}</p>
            )}
          </div>
        </div>

        {followers != null && (
          <div className="flex items-center gap-1.5 text-sm">
            <IconUsers className="text-muted-foreground size-4" />
            <span className="font-medium">{followers.toLocaleString()}</span>
            <span className="text-muted-foreground">followers</span>
          </div>
        )}

        <div className="flex gap-2">
          {isSubscribed ? (
            <Button asChild variant="outline" size="sm" className="flex-1">
              <a
                href={page.details?.link ?? "#"}
                target="_blank"
                rel="noopener noreferrer"
              >
                <IconExternalLink className="size-3.5" />
                View Page
              </a>
            </Button>
          ) : (
            <Button
              size="sm"
              className="flex-1"
              onClick={() => setConfirmOpen(true)}
              disabled={subscribe.isPending}
            >
              {subscribe.isPending ? (
                <IconLoader2 className="size-3.5 animate-spin" />
              ) : (
                <IconBrandMessenger className="size-3.5" />
              )}
              Connect
            </Button>
          )}
        </div>

        {isSubscribed && page.webhook_subscribed_at && (
          <p className="text-muted-foreground flex items-center gap-1 text-xs">
            <IconCheck className="text-success size-3" />
            Subscribed{" "}
            {new Date(page.webhook_subscribed_at).toLocaleDateString()}
          </p>
        )}
      </div>

      <ConfirmDeleteDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title="Enable Messenger for this page?"
        description={`This will subscribe "${name}" to receive Messenger messages in your workspace.`}
        confirmText="Connect"
        onConfirm={() => {
          subscribe.mutate(page.platform_account_id ?? page.id, {
            onSuccess: () => setConfirmOpen(false),
          });
        }}
        loading={subscribe.isPending}
      />
    </>
  );
}

export function MessengerConnection() {
  const params = useParams<{ id: string }>();
  const { data, isLoading } = useMessengerConnectionStatus();

  const status = data?.data;

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
          <IconBrandMessenger className="size-10 text-primary" />
        </div>
        <div className="space-y-1">
          <h2 className="text-xl font-semibold">Connect Messenger</h2>
          <p className="text-muted-foreground max-w-sm text-sm">
            Messenger uses your Facebook Pages. Connect Facebook first to enable
            Messenger in your workspace.
          </p>
        </div>
        <Button asChild size="lg">
          <Link href={`/${params.id}/connections/facebook`}>
            <IconBrandFacebook className="size-4" />
            Connect Facebook
          </Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <div className="flex size-10 items-center justify-center rounded-full bg-primary/10">
          <IconBrandMessenger className="size-5 text-primary" />
        </div>
        <div>
          <h2 className="font-semibold">Messenger Pages</h2>
          <p className="text-muted-foreground text-sm">
            {status.pages.length} page{status.pages.length !== 1 ? "s" : ""}{" "}
            available
          </p>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {status.pages.map((page) => (
          <MessengerPageCard key={page.id} page={page} />
        ))}
      </div>
    </div>
  );
}
