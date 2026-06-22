"use client";

import { useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import {
  IconAlertTriangle,
  IconBolt,
  IconBrandFacebook,
  IconBrandInstagram,
  IconCalendar,
  IconCheck,
  IconLink,
  IconLoader2,
  IconLock,
  IconMessage,
  IconSparkles,
  IconUser,
} from "@tabler/icons-react";
import { Badge } from "@repo/ui/components/ui/badge";
import { Button } from "@repo/ui/components/ui/button";
import { Skeleton } from "@repo/ui/components/ui/skeleton";
import {
  useActivateInstagram,
  useBusinessAppConnections,
  useInstagramConnectionStatus,
  useSubscribeInstagramWebhook,
} from "../client/connections";
import { getInstagramConnectURL } from "../api/connections";
import type { ConnectedPage } from "../types/connections";

// Messaging lives inside the account card as a section (not a separate card).
// Until the account is activated it's locked; once active it shows the
// Facebook-Page requirement and the "Enable messaging" action.
function MessagingSection({
  account,
  active,
}: {
  account: ConnectedPage;
  active: boolean;
}) {
  const subscribe = useSubscribeInstagramWebhook();
  const isSubscribed = !!account.webhook_subscribed_at;
  const subscribedAt = account.webhook_subscribed_at
    ? new Date(account.webhook_subscribed_at).toLocaleDateString()
    : null;

  // Locked state — account not yet activated, so messaging can't be enabled.
  if (!active) {
    return (
      <div className="bg-muted/40 flex items-start gap-3 rounded-lg border border-dashed p-4">
        <div className="bg-muted flex size-9 shrink-0 items-center justify-center rounded-full">
          <IconLock className="text-muted-foreground size-5" />
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium">Messaging</p>
          <p className="text-muted-foreground text-sm">
            Connect with the app above to unlock direct messages.
          </p>
        </div>
      </div>
    );
  }

  // Active + already subscribed — messaging is live.
  if (isSubscribed) {
    return (
      <div className="border-success/30 bg-success/10 flex items-start gap-3 rounded-lg border p-4">
        <div className="bg-success/15 flex size-9 shrink-0 items-center justify-center rounded-full">
          <IconCheck className="text-success size-5" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="font-semibold">Messaging active</p>
            <Badge
              variant="secondary"
              className="bg-success/15 text-success shrink-0"
            >
              Receiving DMs
            </Badge>
          </div>
          <p className="text-muted-foreground text-sm">
            Direct messages are delivered to your inbox
            {subscribedAt ? ` since ${subscribedAt}` : ""}.
          </p>
        </div>
      </div>
    );
  }

  // Active but messaging not yet enabled.
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-start gap-3">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-warning/15">
          <IconMessage className="text-warning size-5" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="font-semibold">Enable messaging</p>
            <Badge
              variant="outline"
              className="border-warning/40 text-warning shrink-0"
            >
              Action needed
            </Badge>
          </div>
          <p className="text-muted-foreground text-sm">
            One more step before Instagram starts sending us your DMs.
          </p>
        </div>
      </div>

      {/* Facebook Page requirement — Instagram only delivers DMs when the
          professional account is linked to a Facebook Page. */}
      <div className="rounded-lg border border-warning/30 bg-warning/10 p-3.5">
        <div className="flex items-center gap-2">
          <IconBrandFacebook className="size-4 shrink-0 text-primary" />
          <IconLink className="text-muted-foreground size-3.5 shrink-0" />
          <IconBrandInstagram className="size-4 shrink-0 text-primary" />
          <p className="text-sm font-medium">
            Link your Facebook Page first
          </p>
        </div>
        <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
          Connect your Facebook Page with your Instagram professional account to
          use messaging. Without that link, Instagram won&apos;t deliver DMs and
          the AI can&apos;t reply.
        </p>
      </div>

      <ul className="flex flex-col gap-2 text-sm">
        <li className="flex items-center gap-2">
          <IconMessage className="size-4 shrink-0 text-primary" />
          <span>Receive customer DMs in your unified inbox</span>
        </li>
        <li className="flex items-center gap-2">
          <IconSparkles className="size-4 shrink-0 text-primary" />
          <span>Let the AI auto-reply and qualify leads</span>
        </li>
        <li className="flex items-center gap-2">
          <IconBolt className="size-4 shrink-0 text-primary" />
          <span>Capture every conversation in real time</span>
        </li>
      </ul>

      <Button
        onClick={() => subscribe.mutate()}
        disabled={subscribe.isPending}
        className="cursor-pointer"
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

function AccountCard({ account }: { account: ConnectedPage }) {
  const activate = useActivateInstagram();
  const isActive = account.is_active;

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
    <div className="flex max-w-xl flex-col gap-5 rounded-xl border p-5">
      <div className="flex items-start gap-4">
        <div className="flex size-12 shrink-0 items-center justify-center rounded-full bg-primary p-[2px]">
          <div className="bg-background flex size-full items-center justify-center rounded-full">
            <IconBrandInstagram className="size-6 text-primary" />
          </div>
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <p className="truncate font-semibold">{name}</p>
            {isActive ? (
              <Badge variant="secondary" className="text-success shrink-0">
                Active
              </Badge>
            ) : (
              <Badge
                variant="outline"
                className="text-muted-foreground shrink-0"
              >
                Not active
              </Badge>
            )}
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

      {/* Activation gate — the credential is stored inactive after OAuth, so the
          user must explicitly connect it with the app before it does anything. */}
      {!isActive && (
        <div className="flex flex-col gap-3 rounded-lg border border-warning/30 bg-warning/10 p-4">
          <div className="flex items-start gap-3">
            <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-warning/15">
              <IconBolt className="text-warning size-5" />
            </div>
            <div className="min-w-0 flex-1">
              <p className="font-semibold">
                Finish connecting
              </p>
              <p className="text-sm text-muted-foreground">
                Your account is linked but not active yet. Connect it with the
                app to start using it.
              </p>
            </div>
          </div>
          <Button
            onClick={() => activate.mutate()}
            disabled={activate.isPending}
            className="cursor-pointer"
          >
            {activate.isPending ? (
              <IconLoader2 className="size-4 animate-spin" />
            ) : (
              <IconBrandInstagram className="size-4" />
            )}
            {activate.isPending ? "Connecting…" : "Connect with app"}
          </Button>
        </div>
      )}

      <div className="border-t pt-5">
        <MessagingSection account={account} active={isActive} />
      </div>
    </div>
  );
}

export function InstagramConnection() {
  const params = useParams<{ id: string }>();
  const { data, isLoading } = useInstagramConnectionStatus();
  const { data: apps, isLoading: appsLoading } = useBusinessAppConnections();
  const [connecting, setConnecting] = useState(false);

  const status = data?.data;
  const facebookConnected = !!apps?.data?.facebook;

  async function handleConnect() {
    setConnecting(true);
    try {
      const url = await getInstagramConnectURL();
      window.location.href = url;
    } finally {
      setConnecting(false);
    }
  }

  if (isLoading || appsLoading) {
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
          <IconBrandInstagram className="size-10 text-primary" />
        </div>
        <div className="space-y-1">
          <h2 className="text-xl font-semibold">Connect Instagram</h2>
          <p className="text-muted-foreground max-w-sm text-sm">
            Link your Instagram Business account to publish content, manage
            comments, and reply to messages from one place.
          </p>
        </div>

        {/* Facebook is a hard prerequisite — Instagram only delivers DMs and
            comment events when the professional account is linked to a Facebook
            Page, so we block connecting until Facebook is set up. */}
        {!facebookConnected && (
          <div className="max-w-md rounded-lg border border-warning/30 bg-warning/10 p-3.5 text-left">
            <div className="flex items-center gap-2">
              <IconAlertTriangle className="text-warning size-4 shrink-0" />
              <p className="text-sm font-medium">
                Connect a Facebook Page first
              </p>
            </div>
            <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
              Without linking your Instagram to a Facebook Page, you won&apos;t be
              able to manage messages, comments, and replies with AI.
            </p>
            <Button
              asChild
              variant="outline"
              size="sm"
              className="border-warning/40 text-warning hover:bg-warning/10 mt-3"
            >
              <Link href={`/${params.id}/connections/facebook`}>
                <IconBrandFacebook className="size-4 text-primary" />
                Connect Facebook
              </Link>
            </Button>
          </div>
        )}

        <Button
          size="lg"
          onClick={handleConnect}
          disabled={connecting || !facebookConnected}
          className="cursor-pointer"
        >
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
        <div className="flex size-10 items-center justify-center rounded-full bg-primary/10">
          <IconBrandInstagram className="size-5 text-primary" />
        </div>
        <div>
          <h2 className="font-semibold">Instagram Account</h2>
          <p className="text-muted-foreground text-sm">
            Your business account is connected
          </p>
        </div>
      </div>

      {status.account && <AccountCard account={status.account} />}
    </div>
  );
}
