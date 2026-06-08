"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  ArrowLeft,
  Check,
  Lock,
  Loader2,
  RefreshCw,
  ShieldCheck,
} from "lucide-react";
import { usePaddle } from "../hooks/usePaddle";
import { useGetPlans } from "../client/plans";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@repo/ui/components/ui/card";
import { Button } from "@repo/ui/components/ui/button";
import { Badge } from "@repo/ui/components/ui/badge";
import { Separator } from "@repo/ui/components/ui/separator";
import { Plan } from "@/lib/types/plans";

function formatLimit(value: number): string {
  if (value === -1) return "Unlimited";
  return value.toLocaleString();
}

const SUMMARY_LIMITS: { key: keyof Plan; label: string }[] = [
  { key: "max_members", label: "Team members" },
  { key: "max_social_accounts", label: "Social accounts" },
  { key: "max_posts_month", label: "Posts / month" },
  { key: "max_ai_replies_month", label: "AI replies / month" },
  { key: "max_leads", label: "Leads" },
];

const SUMMARY_FEATURES: { key: keyof Plan; label: string }[] = [
  { key: "has_post_analytics", label: "Post analytics" },
  { key: "has_multi_platform_post", label: "Multi-platform posting" },
  { key: "has_video_upload", label: "Video upload" },
  { key: "has_ai_dm_replies", label: "AI DM replies" },
  { key: "has_ai_comment_replies", label: "AI comment replies" },
  { key: "has_bookings", label: "Bookings" },
  { key: "has_inventory", label: "Inventory management" },
  { key: "has_priority_support", label: "Priority support" },
];

function CheckoutSkeleton() {
  return (
    <div className="min-h-screen bg-background px-4 py-12">
      <div className="mx-auto max-w-lg">
        <div className="h-5 w-28 animate-pulse rounded bg-muted" />
        <div className="mt-8 h-7 w-44 animate-pulse rounded bg-muted" />
        <div className="mt-1 h-4 w-64 animate-pulse rounded bg-muted" />
        <Card className="mt-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="h-5 w-32 animate-pulse rounded bg-muted" />
              <div className="h-5 w-16 animate-pulse rounded-full bg-muted" />
            </div>
            <div className="mt-3 h-9 w-24 animate-pulse rounded bg-muted" />
          </CardHeader>
          <CardContent className="space-y-3">
            {Array.from({ length: 9 }).map((_, i) => (
              <div
                key={i}
                className="h-4 w-full animate-pulse rounded bg-muted"
              />
            ))}
            <div className="h-10 w-full animate-pulse rounded-4xl bg-muted" />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export function CheckoutClient() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const planId = searchParams.get("plan");
  const billing = (searchParams.get("billing") ?? "monthly") as
    | "monthly"
    | "yearly";

  const paddle = usePaddle();
  const { data, isLoading } = useGetPlans();
  const plans = data?.data ?? [];
  const plan = plans.find((p) => p.id === planId);

  if (isLoading) return <CheckoutSkeleton />;

  if (!plan) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">Plan not found.</p>
        <Button variant="outline" asChild>
          <Link href="/plans">Back to plans</Link>
        </Button>
      </div>
    );
  }

  const price =
    billing === "monthly" ? plan.monthly_price : plan.yearly_price;
  const isFree = price === 0;
  const priceId =
    billing === "monthly"
      ? plan.paddle_price_id_monthly
      : plan.paddle_price_id_yearly;

  const monthlySaving =
    billing === "yearly" && plan.monthly_price > 0
      ? (plan.monthly_price * 12 - plan.yearly_price) / 100
      : 0;

  const enabledFeatures = SUMMARY_FEATURES.filter(
    ({ key }) => plan[key] as boolean,
  );

  const handleCheckout = () => {
    if (!paddle || !priceId) return;
    paddle.Checkout.open({
      items: [{ priceId, quantity: 1 }],
      settings: {
        successUrl: `${process.env.NEXT_PUBLIC_BETTER_AUTH_URL}/checkout/status?success=true`,
      },
    });
  };

  return (
    <div className="min-h-screen bg-background px-4 py-12">
      <div className="mx-auto max-w-lg">
        {/* Back */}
        <Link
          href="/plans"
          className="inline-flex cursor-pointer items-center gap-1.5 text-sm text-muted-foreground transition-colors duration-200 hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          Back to plans
        </Link>

        <div className="mt-8">
          <h1 className="text-2xl font-semibold tracking-tight">
            Order summary
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Review your selection before proceeding to payment.
          </p>
        </div>

        <Card className="mt-6">
          <CardHeader className="pb-4">
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg capitalize">{plan.name}</CardTitle>
              <Badge variant={billing === "yearly" ? "default" : "secondary"}>
                {billing === "yearly" ? "Yearly" : "Monthly"}
              </Badge>
            </div>

            <div className="mt-3 flex items-end gap-1">
              <span className="text-3xl font-bold tracking-tight">
                {isFree ? "Free" : `$${(price / 100).toFixed(2)}`}
              </span>
              {!isFree && (
                <span className="mb-1 text-sm text-muted-foreground">
                  / {billing === "monthly" ? "mo" : "yr"}
                </span>
              )}
            </div>

            {billing === "yearly" && !isFree && monthlySaving > 0 && (
              <p className="text-xs font-medium text-primary">
                ${(plan.yearly_price / 100 / 12).toFixed(2)}/mo billed
                annually &middot; saves ${monthlySaving.toFixed(2)}/yr
              </p>
            )}
          </CardHeader>

          <CardContent className="space-y-4">
            <div className="space-y-2">
              {SUMMARY_LIMITS.map(({ key, label }) => (
                <div
                  key={key}
                  className="flex items-center justify-between text-sm"
                >
                  <span className="text-muted-foreground">{label}</span>
                  <span className="font-medium">
                    {formatLimit(plan[key] as number)}
                  </span>
                </div>
              ))}
            </div>

            {enabledFeatures.length > 0 && (
              <>
                <Separator />
                <div className="space-y-2">
                  {enabledFeatures.map(({ key, label }) => (
                    <div key={key} className="flex items-center gap-2 text-sm">
                      <Check className="size-3.5 shrink-0 text-primary" />
                      <span>{label}</span>
                    </div>
                  ))}
                </div>
              </>
            )}

            <Separator />

            <div className="flex items-center justify-between text-sm font-semibold">
              <span>Total today</span>
              <span>{isFree ? "Free" : `$${(price / 100).toFixed(2)}`}</span>
            </div>

            {isFree ? (
              <Button
                className="w-full cursor-pointer"
                size="lg"
                onClick={() => router.push("/onboard")}
              >
                Get started free
              </Button>
            ) : (
              <Button
                className="w-full cursor-pointer"
                size="lg"
                disabled={!paddle || !priceId}
                onClick={handleCheckout}
              >
                {!paddle ? (
                  <>
                    <Loader2 className="size-4 animate-spin" />
                    Loading...
                  </>
                ) : !priceId ? (
                  "Not available"
                ) : (
                  "Continue to payment"
                )}
              </Button>
            )}

            {!isFree && (
              <div className="flex flex-wrap items-center justify-center gap-x-4 gap-y-1.5 text-xs text-muted-foreground">
                <span className="flex items-center gap-1">
                  <Lock className="size-3" />
                  Secure payment
                </span>
                <span className="flex items-center gap-1">
                  <RefreshCw className="size-3" />
                  Cancel anytime
                </span>
                <span className="flex items-center gap-1">
                  <ShieldCheck className="size-3" />
                  Powered by Paddle
                </span>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
