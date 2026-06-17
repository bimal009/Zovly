"use client";

import { useState } from "react";
import {
  IconBrandInstagram,
  IconCalendar,
  IconLoader2,
  IconUser,
} from "@tabler/icons-react";
import { Badge } from "@repo/ui/components/ui/badge";
import { Button } from "@repo/ui/components/ui/button";
import { Skeleton } from "@repo/ui/components/ui/skeleton";
import { getInstagramConnectURL } from "../api/connections";
import { useInstagramConnectionStatus } from "../client/connections";
import type { ConnectedPage } from "../types/connections";

function AccountCard({ account }: { account: ConnectedPage }) {
  const name =
    account.platform_account_name ?? account.platform_account_id ?? "Instagram Account";
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
        <div className="max-w-sm">
          <AccountCard account={status.account} />
        </div>
      )}
    </div>
  );
}
