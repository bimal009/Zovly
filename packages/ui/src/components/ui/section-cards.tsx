import * as React from "react";
import { IconTrendingDown, IconTrendingUp } from "@tabler/icons-react";

import { Badge } from "./badge";
import {
  Card,
  CardAction,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "./card";

// ─── Types ────────────────────────────────────────────────────────────────────

export type StatCardItem = {
  label: string;
  value: string | number;
  trend?: string;
  trendUp?: boolean;
  description?: string;
  sub?: string;
  icon?: React.ComponentType<{ className?: string }>;
};

export type SectionCardsProps = {
  cards: StatCardItem[];
  cols?: 2 | 3 | 4;
};

// ─── Grid column mapping ──────────────────────────────────────────────────────

const colClass: Record<2 | 3 | 4, string> = {
  2: "@xl/main:grid-cols-2",
  3: "@xl/main:grid-cols-2 @4xl/main:grid-cols-3",
  4: "@xl/main:grid-cols-2 @5xl/main:grid-cols-4",
};

// ─── Component ────────────────────────────────────────────────────────────────

export function SectionCards({ cards, cols = 4 }: SectionCardsProps) {
  return (
    <div
      className={[
        "grid grid-cols-1 gap-4 px-4 lg:px-6",
        "*:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:from-primary/5 *:data-[slot=card]:to-card *:data-[slot=card]:shadow-xs",
        "dark:*:data-[slot=card]:bg-card",
        colClass[cols],
      ].join(" ")}
    >
      {cards.map((card) => (
        <StatCard key={card.label} card={card} />
      ))}
    </div>
  );
}

// ─── Individual card ──────────────────────────────────────────────────────────

function StatCard({ card }: { card: StatCardItem }) {
  const Icon = card.icon;
  const hasTrend = card.trend !== undefined;
  const trendUp = card.trendUp ?? true;

  return (
    <Card className="@container/card">
      <CardHeader>
        <CardDescription className="flex items-center gap-1.5">
          {Icon && <Icon className="size-3.5 shrink-0" />}
          {card.label}
        </CardDescription>
        <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
          {card.value}
        </CardTitle>
        {hasTrend && (
          <CardAction>
            <Badge
              variant="outline"
              className={
                trendUp
                  ? "border-green-200 bg-green-50 text-green-700 dark:border-green-800 dark:bg-green-950/50 dark:text-green-400"
                  : "border-red-200 bg-red-50 text-red-700 dark:border-red-800 dark:bg-red-950/50 dark:text-red-400"
              }
            >
              {trendUp ? (
                <IconTrendingUp className="size-3" />
              ) : (
                <IconTrendingDown className="size-3" />
              )}
              {card.trend}
            </Badge>
          </CardAction>
        )}
      </CardHeader>
      {(card.description || card.sub) && (
        <CardFooter className="flex-col items-start gap-1 text-sm">
          {card.description && (
            <div className="line-clamp-1 flex items-center gap-1.5 font-medium">
              {hasTrend &&
                (trendUp ? (
                  <IconTrendingUp className="size-3.5 shrink-0 text-green-600 dark:text-green-400" />
                ) : (
                  <IconTrendingDown className="size-3.5 shrink-0 text-red-500 dark:text-red-400" />
                ))}
              {card.description}
            </div>
          )}
          {card.sub && (
            <div className="text-muted-foreground">{card.sub}</div>
          )}
        </CardFooter>
      )}
    </Card>
  );
}
