"use client";

import Link from "next/link";
import { Zap } from "lucide-react";
import { useSession } from "@/lib/auth-client";
import { Button } from "@repo/ui/components/ui/button";

export function Navbar() {
  const { data: sessionData, isPending: sessionPending } = useSession();
  console.log(sessionData?.session.token);
  const user = sessionData?.user;
  const hasSession = !!sessionData?.session;

  let ctaHref = "/signup";
  let ctaLabel = "Get started free";
  let showSignIn = true;
  const businessId = "495f97b5-b065-46b1-9289-e660b4fbea29";

  if (hasSession) {
    showSignIn = false;
    if (businessId) {
      ctaHref = `/${businessId}/dashboard`;
      ctaLabel = "Dashboard";
    } else {
      ctaHref = "/onboard";
      ctaLabel = "Get Started";
    }
  }

  const loading = !!sessionPending;

  return (
    <header className="sticky top-0 z-50 border-b border-border/50 bg-background/80 backdrop-blur-md">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4 lg:px-6">
        {/* Brand */}
        <Link
          href="/"
          className="flex items-center gap-2 text-foreground transition-opacity hover:opacity-75"
          aria-label="SocialOS home"
        >
          <Zap className="h-5 w-5 text-primary" aria-hidden="true" />
          <span className="text-base font-semibold tracking-tight">
            SocialOS
          </span>
        </Link>

        {/* Actions */}
        <nav className="flex items-center gap-2" aria-label="Site navigation">
          {hasSession && user?.email && (
            <span className="hidden text-sm text-muted-foreground sm:block">
              {user.email}
            </span>
          )}
          {showSignIn && (
            <Button variant="ghost" size="sm" asChild>
              <Link href="/signin">Sign in</Link>
            </Button>
          )}
          <Button size="sm" disabled={loading} asChild={!loading}>
            {loading ? (
              <span>{ctaLabel}</span>
            ) : (
              <Link href={ctaHref}>{ctaLabel}</Link>
            )}
          </Button>
        </nav>
      </div>
    </header>
  );
}
