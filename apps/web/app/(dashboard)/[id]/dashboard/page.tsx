import { SiteHeader } from "@repo/ui/components/ui/site-header";
import { ChartAreaInteractive } from "@repo/ui/components/ui/chart-area-interactive";
import { ChartBarProducts } from "@repo/ui/components/ui/chart-bar-products";
import { OverviewCards } from "@/features/dashboard/components/overview-cards";
import { AdCampaigns } from "@/features/dashboard/components/ad-campaigns";
import { RecentOrders } from "@/features/dashboard/components/recent-orders";

export default function DashboardPage() {
  return (
    <>
      <SiteHeader title="Overview" />
      <div className="flex flex-1 flex-col">
        <div className="@container/main flex flex-1 flex-col gap-2">
          <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
            {/* KPI Cards */}
            <OverviewCards />

            {/* Reach / Engagement chart — full width */}
            <div className="px-4 lg:px-6">
              <ChartAreaInteractive />
            </div>

            {/* Bar chart (top products) + Ad campaigns — two columns */}
            <div className="grid gap-4 px-4 lg:px-6 @xl/main:grid-cols-5">
              <div className="@xl/main:col-span-3">
                <ChartBarProducts />
              </div>
              <div className="@xl/main:col-span-2">
                <AdCampaigns />
              </div>
            </div>

            {/* Recent AI-captured leads */}
            <RecentOrders />
          </div>
        </div>
      </div>
    </>
  );
}
