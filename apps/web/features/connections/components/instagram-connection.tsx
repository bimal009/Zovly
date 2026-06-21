"use client";

import { useState } from "react";
import {
  IconBolt,
  IconBrandInstagram,
  IconCalendar,
  IconCheck,
  IconLoader2,
  IconMessage,
  IconSparkles,
  IconUser,
} from "@tabler/icons-react";
import { Badge } from "@repo/ui/components/ui/badge";
import { Button } from "@repo/ui/components/ui/button";
import { Skeleton } from "@repo/ui/components/ui/skeleton";
import {
  useInstagramConnectionStatus,
  useSubscribeInstagramWebhook,
} from "../client/connections";
import { getInstagramConnectURL } from "../api/connections";
import type { ConnectedPage } from "../types/connections";

function AccountCard({ account }: { account: ConnectedPage }) {
  const name =
    account.platform_account_name ??
    account.platform_account_id ??
    "Instagram Account";
  const expiresAt = account.token_expires_at
    ? new Date(account.token_expires_at).toLocaleDateString()
    : null;
  const connectedAt = account.connected_at
    ? new Date(account.connected_at).toLocaleDateString()
    : null;

  return (
    <div className="flex flex-col gap-4 rounded-xl border p-5">
      <div className="flex items-start gap-4">
        <div className="flex size-12 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-[#833AB4] via-[#FD1D1D] to-[#FCAF45] p-[2px]">
          <div className="flex size-full items-center justify-center rounded-full bg-background">
            <IconBrandInstagram className="size-6 text-[#E1306C]" />
          </div>
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <p className="truncate font-semibold">{name}</p>
            <Badge variant="secondary" className="shrink-0 text-green-600">
              Connected
            </Badge>
          </div>
          {account.platform_account_id && (
            <p className="text-muted-foreground text-xs">
              ID: {account.platform_account_id}
            </p>
          )}
        </div>
      </div>

      <div className="flex flex-col gap-1.5 text-sm">
        {connectedAt && (
          <div className="flex items-center gap-1.5">
            <IconUser className="text-muted-foreground size-4" />
            <span className="text-muted-foreground">Connected on</span>
            <span className="font-medium">{connectedAt}</span>
          </div>
        )}
        {expiresAt && (
          <div className="flex items-center gap-1.5">
            <IconCalendar className="text-muted-foreground size-4" />
            <span className="text-muted-foreground">Token expires</span>
            <span className="font-medium">{expiresAt}</span>
          </div>
        )}
      </div>
    </div>
  );
}

function MessagingCard({ account }: { account: ConnectedPage }) {
  const subscribe = useSubscribeInstagramWebhook();
  const isSubscribed = !!account.webhook_subscribed_at;
  const subscribedAt = account.webhook_subscribed_at
    ? new Date(account.webhook_subscribed_at).toLocaleDateString()
    : null;

  if (isSubscribed) {
    return (
      <div className="flex flex-col gap-3 rounded-xl border border-green-200 bg-green-50 p-5 dark:border-green-900/40 dark:bg-green-950/20">
        <div className="flex items-start gap-3">
          <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/40">
            <IconCheck className="size-5 text-green-600 dark:text-green-400" />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <p className="font-semibold text-green-900 dark:text-green-100">
                Messaging active
              </p>
              <Badge
                variant="secondary"
                className="shrink-0 bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300"
              >
                Receiving DMs
              </Badge>
            </div>
            <p className="text-sm text-green-800/80 dark:text-green-200/70">
              Direct messages are being delivered to your inbox
              {subscribedAt ? ` since ${subscribedAt}` : ""}.
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 rounded-xl border border-amber-200 bg-amber-50 p-5 dark:border-amber-900/40 dark:bg-amber-950/20">
      <div className="flex items-start gap-3">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-amber-100 dark:bg-amber-900/40">
          <IconMessage className="size-5 text-amber-600 dark:text-amber-400" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="font-semibold text-amber-900 dark:text-amber-100">
              Enable message receiving
            </p>
            <Badge
              variant="outline"
              className="shrink-0 border-amber-300 text-amber-700 dark:border-amber-800 dark:text-amber-300"
            >
              Action needed
            </Badge>
          </div>
          <p className="text-sm text-amber-800/90 dark:text-amber-200/80">
            Your account is linked, but Instagram needs one more step before it
            will send us your DMs.
          </p>
        </div>
      </div>

      <div className="rounded-lg border border-amber-200/70 bg-background/60 p-4 dark:border-amber-900/30">
        <p className="text-sm font-medium">Why is this needed?</p>
        <p className="text-muted-foreground mt-1 text-sm leading-relaxed">
          Connecting grants access to your profile, but Instagram requires a
          separate <span className="font-medium">webhook subscription</span>{" "}
          before it delivers direct messages. Until you enable it, incoming DMs
          never reach Zovly and the AI can&apos;t reply.
        </p>
        <ul className="mt-3 flex flex-col gap-2 text-sm">
          <li className="flex items-center gap-2">
            <IconMessage className="size-4 shrink-0 text-[#E1306C]" />
            <span>Receive customer DMs in your unified inbox</span>
          </li>
          <li className="flex items-center gap-2">
            <IconSparkles className="size-4 shrink-0 text-[#E1306C]" />
            <span>Let the AI auto-reply and qualify leads</span>
          </li>
          <li className="flex items-center gap-2">
            <IconBolt className="size-4 shrink-0 text-[#E1306C]" />
            <span>Capture every conversation in real time</span>
          </li>
        </ul>
      </div>

      <Button
        onClick={() => subscribe.mutate()}
        disabled={subscribe.isPending}
        className="bg-[#E1306C] hover:bg-[#E1306C]/90"
      >
        {subscribe.isPending ? (
          <IconLoader2 className="size-4 animate-spin" />
        ) : (
          <IconMessage className="size-4" />
        )}
        {subscribe.isPending ? "Enabling…" : "Enable messaging"}
      </Button>
    </div>
  );
}

export function InstagramConnection() {
  const { data, isLoading } = useInstagramConnectionStatus();
  const [connecting, setConnecting] = useState(false);

  const status = data?.data;

  async function handleConnect() {
    setConnecting(true);
    try {
      const url = await getInstagramConnectURL();
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
        <div className="flex size-20 items-center justify-center rounded-full bg-gradient-to-br from-[#833AB4]/10 via-[#FD1D1D]/10 to-[#FCAF45]/10">
          <IconBrandInstagram className="size-10 text-[#E1306C]" />
        </div>
        <div className="space-y-1">
          <h2 className="text-xl font-semibold">Connect Instagram</h2>
          <p className="text-muted-foreground max-w-sm text-sm">
            Link your Instagram Business account to publish content, manage
            comments, and reply to messages from one place.
          </p>
        </div>
        <Button size="lg" onClick={handleConnect} disabled={connecting}>
          {connecting ? (
            <IconLoader2 className="size-4 animate-spin" />
          ) : (
            <IconBrandInstagram className="size-4" />
          )}
          {connecting ? "Redirecting…" : "Connect Now"}
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <div className="flex size-10 items-center justify-center rounded-full bg-gradient-to-br from-[#833AB4]/10 via-[#FD1D1D]/10 to-[#FCAF45]/10">
          <IconBrandInstagram className="size-5 text-[#E1306C]" />
        </div>
        <div>
          <h2 className="font-semibold">Instagram Account</h2>
          <p className="text-muted-foreground text-sm">
            Your business account is connected
          </p>
        </div>
      </div>

      {status.account && (
        <div className="grid max-w-3xl gap-4 sm:grid-cols-2">
          <AccountCard account={status.account} />
          <MessagingCard account={status.account} />
        </div>
      )}
    </div>
  );
}
