"use client";

import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis } from "recharts";

import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "./chart";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./select";
import { ToggleGroup, ToggleGroupItem } from "./toggle-group";
import { useIsMobile } from "../../../hooks/use-mobile";

const chartData = [
  { date: "2026-03-15", reach: 18400, engagement: 920 },
  { date: "2026-03-22", reach: 22100, engagement: 1105 },
  { date: "2026-03-29", reach: 19800, engagement: 990 },
  { date: "2026-04-05", reach: 31200, engagement: 1560 },
  { date: "2026-04-12", reach: 28900, engagement: 1445 },
  { date: "2026-04-19", reach: 35600, engagement: 1780 },
  { date: "2026-04-26", reach: 41200, engagement: 2060 },
  { date: "2026-05-03", reach: 38700, engagement: 1935 },
  { date: "2026-05-10", reach: 52300, engagement: 2615 },
  { date: "2026-05-17", reach: 61800, engagement: 3090 },
  { date: "2026-05-24", reach: 57400, engagement: 2870 },
  { date: "2026-05-31", reach: 74900, engagement: 3745 },
  { date: "2026-06-07", reach: 83200, engagement: 4160 },
];

const chartConfig = {
  reach: {
    label: "Reach",
    color: "var(--primary)",
  },
  engagement: {
    label: "Engagement",
    color: "var(--muted-foreground)",
  },
} satisfies ChartConfig;

export function ChartAreaInteractive() {
  const isMobile = useIsMobile();
  const [timeRange, setTimeRange] = React.useState("90d");

  React.useEffect(() => {
    if (isMobile) {
      setTimeRange("30d");
    }
  }, [isMobile]);

  const filteredData = chartData.filter((item) => {
    const date = new Date(item.date);
    const referenceDate = new Date("2026-06-12");
    let daysToSubtract = 90;
    if (timeRange === "30d") {
      daysToSubtract = 30;
    } else if (timeRange === "7d") {
      daysToSubtract = 7;
    }
    const startDate = new Date(referenceDate);
    startDate.setDate(startDate.getDate() - daysToSubtract);
    return date >= startDate;
  });

  return (
    <Card className="@container/card">
      <CardHeader>
        <CardTitle>Cross-Platform Performance</CardTitle>
        <CardDescription>
          <span className="hidden @[540px]/card:block">
            Reach vs Engagement across all connected platforms
          </span>
          <span className="@[540px]/card:hidden">Reach vs Engagement</span>
        </CardDescription>
        <CardAction>
          <ToggleGroup
            type="single"
            value={timeRange}
            onValueChange={setTimeRange}
            variant="outline"
            className="hidden *:data-[slot=toggle-group-item]:px-4! @[767px]/card:flex"
          >
            <ToggleGroupItem value="90d">Last 3 months</ToggleGroupItem>
            <ToggleGroupItem value="30d">Last 30 days</ToggleGroupItem>
            <ToggleGroupItem value="7d">Last 7 days</ToggleGroupItem>
          </ToggleGroup>
          <Select value={timeRange} onValueChange={setTimeRange}>
            <SelectTrigger
              className="flex w-40 **:data-[slot=select-value]:block **:data-[slot=select-value]:truncate @[767px]/card:hidden"
              size="sm"
              aria-label="Select a value"
            >
              <SelectValue placeholder="Last 3 months" />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              <SelectItem value="90d" className="rounded-lg">
                Last 3 months
              </SelectItem>
              <SelectItem value="30d" className="rounded-lg">
                Last 30 days
              </SelectItem>
              <SelectItem value="7d" className="rounded-lg">
                Last 7 days
              </SelectItem>
            </SelectContent>
          </Select>
        </CardAction>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        <ChartContainer
          config={chartConfig}
          className="aspect-auto h-[250px] w-full"
        >
          <AreaChart data={filteredData}>
            <defs>
              <linearGradient id="fillReach" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-reach)"
                  stopOpacity={1.0}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-reach)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id="fillEngagement" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-engagement)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-engagement)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
              tickFormatter={(value) => {
                const date = new Date(value);
                return date.toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                });
              }}
            />
            <ChartTooltip
              cursor={false}
              content={
                <ChartTooltipContent
                  labelFormatter={(value) => {
                    return new Date(value).toLocaleDateString("en-US", {
                      month: "short",
                      day: "numeric",
                    });
                  }}
                  indicator="dot"
                />
              }
            />
            <Area
              dataKey="engagement"
              type="natural"
              fill="url(#fillEngagement)"
              stroke="var(--color-engagement)"
              stackId="a"
            />
            <Area
              dataKey="reach"
              type="natural"
              fill="url(#fillReach)"
              stroke="var(--color-reach)"
              stackId="a"
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
