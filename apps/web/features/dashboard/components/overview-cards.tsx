import {
  SectionCards,
  type StatCardItem,
} from "@repo/ui/components/ui/section-cards";

const metrics: StatCardItem[] = [
  {
    label: "Total Reach",
    value: "245,800",
    trend: "+18.4%",
    trendUp: true,
    description: "Cross-platform this week",
    sub: "Instagram, TikTok, Facebook, YouTube",
  },
  {
    label: "Total Engagement",
    value: "12,430",
    trend: "+5.1%",
    trendUp: true,
    description: "Likes, comments & shares",
    sub: "5.2% avg engagement rate",
  },
  {
    label: "New Leads",
    value: "47",
    trend: "+12",
    trendUp: true,
    description: "AI-captured this week",
    sub: "From DMs & comments",
  },
  {
    label: "Active Campaigns",
    value: "5",
    trend: "-1",
    trendUp: false,
    description: "3 running, 2 paused",
    sub: "Across Meta & TikTok",
  },
];

export function OverviewCards() {
  return <SectionCards cards={metrics} cols={4} />;
}
