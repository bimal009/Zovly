"use client";

import { useState } from "react";
import Link from "next/link";
import { Check, X, Zap } from "lucide-react";
import { useGetPlans } from "../client/plans";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@repo/ui/components/ui/card";
import { Button } from "@repo/ui/components/ui/button";
import { Badge } from "@repo/ui/components/ui/badge";
import { Separator } from "@repo/ui/components/ui/separator";
import { cn } from "@repo/ui/utils";
import { Plan } from "@/lib/types/plans";

function formatLimit(value: number): string {
  if (value === -1) return "Unlimited";
  return value.toLocaleString();
}

const LIMIT_FEATURES: { key: keyof Plan; label: string }[] = [
  { key: "max_members", label: "Team members" },
  { key: "max_social_accounts", label: "Social accounts" },
  { key: "max_posts_month", label: "Posts / month" },
  { key: "max_ai_replies_month", label: "AI replies / month" },
  { key: "max_leads", label: "Leads" },
];

const BOOL_FEATURES: { key: keyof Plan; label: string }[] = [
  { key: "has_post_analytics", label: "Post analytics" },
  { key: "has_multi_platform_post", label: "Multi-platform posting" },
  { key: "has_video_upload", label: "Video upload" },
  { key: "has_ai_dm_replies", label: "AI DM replies" },
  { key: "has_ai_comment_replies", label: "AI comment replies" },
  { key: "has_ai_lead_scoring", label: "AI lead scoring" },
  { key: "has_ai_ad_suggestions", label: "AI ad suggestions" },
  { key: "has_voice_transcription", label: "Voice transcription" },
  { key: "has_image_understanding", label: "Image understanding" },
  { key: "has_bookings", label: "Bookings" },
  { key: "has_inventory", label: "Inventory management" },
  { key: "has_payments", label: "payment Gateways" },
  { key: "has_meta_ads", label: "Meta Ads" },
  { key: "has_tiktok_ads", label: "TikTok Ads" },
  { key: "has_google_workspace", label: "Google Workspace" },
  { key: "has_priority_support", label: "Priority support" },
];

function PlanCardSkeleton() {
  return (
    <Card className="flex flex-col animate-pulse">
      <CardHeader>
        <div className="h-4 w-20 rounded bg-muted" />
        <div className="mt-3 h-9 w-28 rounded bg-muted" />
        <div className="mt-1 h-3.5 w-40 rounded bg-muted" />
      </CardHeader>
      <CardContent className="flex-1 space-y-2.5">
        {Array.from({ length: 10 }).map((_, i) => (
          <div key={i} className="h-4 w-full rounded bg-muted" />
        ))}
      </CardContent>
      <CardFooter>
        <div className="h-10 w-full rounded-4xl bg-muted" />
      </CardFooter>
    </Card>
  );
}

const Plans = () => {
  const [billing, setBilling] = useState<"monthly" | "yearly">("monthly");
  const { data, isLoading, isError } = useGetPlans();

  const plans = data?.data ?? [];

  return (
    <section className="px-4 py-16">
      <div className="mx-auto max-w-6xl">
        {/* Header */}
        <div className="mb-12 text-center">
          <h2 className="text-3xl font-semibold tracking-tight">
            Simple, transparent pricing
          </h2>
          <p className="mt-2 text-base text-muted-foreground">
            Choose the plan that fits your business. Upgrade or downgrade any
            time.
          </p>

          {/* Billing toggle */}
          <div className="mt-6 inline-flex items-center gap-1 rounded-4xl bg-muted p-1">
            <button
              onClick={() => setBilling("monthly")}
              className={cn(
                "cursor-pointer rounded-4xl px-4 py-1.5 text-sm font-medium transition-all duration-200",
                billing === "monthly"
                  ? "bg-background text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              Monthly
            </button>
            <button
              onClick={() => setBilling("yearly")}
              className={cn(
                "flex cursor-pointer items-center gap-1.5 rounded-4xl px-4 py-1.5 text-sm font-medium transition-all duration-200",
                billing === "yearly"
                  ? "bg-background text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              Yearly
              <Badge
                variant="secondary"
                className="h-4 px-1.5 text-[10px] font-semibold"
              >
                Save 20%
              </Badge>
            </button>
          </div>
        </div>

        {/* Plan cards */}
        {isError ? (
          <p className="py-12 text-center text-muted-foreground">
            Failed to load plans. Please try again later.
          </p>
        ) : (
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
            {isLoading
              ? Array.from({ length: 3 }).map((_, i) => (
                  <PlanCardSkeleton key={i} />
                ))
              : plans.map((plan) => {
                  const isPopular = plan.name.toLowerCase().includes("growth");
                  const price =
                    billing === "monthly"
                      ? plan.monthly_price
                      : plan.yearly_price;
                  const isFree = price === 0;

                  return (
                    <Card
                      key={plan.id}
                      className={cn(
                        "flex flex-col transition-all duration-200",
                        isPopular &&
                          "scale-[1.02] shadow-lg shadow-primary/10 ring-2 ring-primary",
                      )}
                    >
                      <CardHeader className="pb-4">
                        <div className="flex items-center justify-between">
                          <CardTitle className="text-base font-semibold capitalize">
                            {plan.name}
                          </CardTitle>
                          {isPopular && (
                            <Badge className="gap-1 text-[11px]">
                              <Zap className="size-3" />
                              Most popular
                            </Badge>
                          )}
                        </div>

                        <div className="mt-3 flex items-end gap-1">
                          <span className="text-3xl font-bold tracking-tight">
                            {isFree ? "Free" : `$${price / 100}`}
                          </span>
                          {!isFree && (
                            <span className="mb-1 text-sm text-muted-foreground">
                              / {billing === "monthly" ? "mo" : "yr"}
                            </span>
                          )}
                        </div>

                        <CardDescription className="min-h-[1.25rem]">
                          {billing === "yearly" && !isFree && (
                            <span className="text-xs font-medium text-primary">
                              ${Math.round(plan.yearly_price / 100 / 12)}/mo
                              billed annually
                            </span>
                          )}
                        </CardDescription>
                      </CardHeader>

                      <CardContent className="flex-1 space-y-3">
                        {/* Numeric limits */}
                        <div className="space-y-2">
                          {LIMIT_FEATURES.map(({ key, label }) => (
                            <div
                              key={key}
                              className="flex items-center justify-between text-sm"
                            >
                              <span className="text-muted-foreground">
                                {label}
                              </span>
                              <span className="font-medium">
                                {formatLimit(plan[key] as number)}
                              </span>
                            </div>
                          ))}
                        </div>

                        <Separator />

                        {/* Boolean features */}
                        <div className="space-y-2">
                          {BOOL_FEATURES.map(({ key, label }) => {
                            const enabled = plan[key] as boolean;
                            return (
                              <div
                                key={key}
                                className={cn(
                                  "flex items-center gap-2 text-sm",
                                  !enabled && "opacity-40",
                                )}
                              >
                                {enabled ? (
                                  <Check className="size-3.5 shrink-0 text-primary" />
                                ) : (
                                  <X className="size-3.5 shrink-0" />
                                )}
                                <span>{label}</span>
                              </div>
                            );
                          })}
                        </div>
                      </CardContent>

                      <CardFooter className="pt-4">
                        <Button
                          className="w-full cursor-pointer"
                          variant={isPopular ? "default" : "outline"}
                          size="lg"
                          asChild
                        >
                          <Link
                            href={
                              isFree
                                ? "/onboard"
                                : `/checkout?plan=${plan.id}&billing=${billing}`
                            }
                          >
                            {isFree ? "Get started free" : "Get started"}
                          </Link>
                        </Button>
                      </CardFooter>
                    </Card>
                  );
                })}
          </div>
        )}
      </div>
    </section>
  );
};

export default Plans;
