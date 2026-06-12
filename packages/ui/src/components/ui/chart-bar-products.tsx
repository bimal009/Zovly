"use client";

import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";

import {
  Card,
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

const data = [
  { name: "Pro Consultation", value: 48 },
  { name: "Starter Package", value: 35 },
  { name: "Blue Jacket (M)", value: 29 },
  { name: "Premium Service", value: 24 },
  { name: "Social Bundle", value: 21 },
  { name: "Basic Plan", value: 18 },
  { name: "Voice Note AI", value: 14 },
];

const chartConfig = {
  value: {
    label: "Sales / Bookings",
    color: "var(--primary)",
  },
} satisfies ChartConfig;

export function ChartBarProducts() {
  return (
    <Card className="@container/card h-full">
      <CardHeader>
        <CardTitle>Top Products &amp; Services</CardTitle>
        <CardDescription>Most sold this month</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[280px] w-full">
          <BarChart
            data={data}
            layout="vertical"
            margin={{ left: 0, right: 16, top: 0, bottom: 0 }}
          >
            <CartesianGrid horizontal={false} />
            <XAxis
              type="number"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              fontSize={12}
            />
            <YAxis
              type="category"
              dataKey="name"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              width={130}
              fontSize={12}
            />
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent hideLabel />}
            />
            <Bar
              dataKey="value"
              fill="var(--color-value)"
              radius={[0, 4, 4, 0]}
            />
          </BarChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
