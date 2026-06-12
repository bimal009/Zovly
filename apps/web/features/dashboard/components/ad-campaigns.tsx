import { Badge } from "@repo/ui/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@repo/ui/components/ui/card";
import { Separator } from "@repo/ui/components/ui/separator";

const adStats = [
  { label: "Total Spend", value: "$2,840", sub: "This month" },
  { label: "Impressions", value: "1.24M", sub: "All campaigns" },
  { label: "Avg CTR", value: "3.8%", sub: "Industry avg 2.1%" },
  { label: "Best ROAS", value: "4.2×", sub: "Summer Sale" },
];

type CampaignStatus = "running" | "paused" | "ended";

const campaigns: {
  name: string;
  platform: string;
  status: CampaignStatus;
  spend: string;
  roas: string;
}[] = [
  {
    name: "Summer Sale 2026",
    platform: "Meta",
    status: "running",
    spend: "$1,240",
    roas: "4.2×",
  },
  {
    name: "New Product Launch",
    platform: "TikTok",
    status: "paused",
    spend: "$890",
    roas: "2.8×",
  },
  {
    name: "Brand Awareness",
    platform: "Meta",
    status: "running",
    spend: "$480",
    roas: "3.1×",
  },
  {
    name: "Retargeting Q2",
    platform: "Google",
    status: "running",
    spend: "$230",
    roas: "5.6×",
  },
];

const statusVariant: Record<
  CampaignStatus,
  "default" | "secondary" | "outline"
> = {
  running: "default",
  paused: "secondary",
  ended: "outline",
};

export function AdCampaigns() {
  return (
    <Card className="@container/card h-full">
      <CardHeader>
        <CardTitle>Ad Campaigns</CardTitle>
        <CardDescription>Active campaign performance</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="grid grid-cols-2 gap-3">
          {adStats.map((stat) => (
            <div key={stat.label} className="rounded-lg border p-3">
              <p className="text-xs text-muted-foreground">{stat.label}</p>
              <p className="mt-0.5 text-xl font-semibold tabular-nums">
                {stat.value}
              </p>
              <p className="text-xs text-muted-foreground">{stat.sub}</p>
            </div>
          ))}
        </div>
        <Separator />
        <div className="flex flex-col gap-3">
          {campaigns.map((c) => (
            <div
              key={c.name}
              className="flex items-center justify-between gap-2"
            >
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{c.name}</p>
                <p className="text-xs text-muted-foreground">
                  {c.platform} · {c.spend}
                </p>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <span className="text-sm font-semibold tabular-nums">
                  {c.roas}
                </span>
                <Badge
                  variant={statusVariant[c.status]}
                  className="capitalize text-xs"
                >
                  {c.status}
                </Badge>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
