import { SiteHeader } from "@repo/ui/components/ui/site-header";
import { ServicesTable } from "@/features/services/components/services-table";

export default function ServicesPage() {
  return (
    <>
      <SiteHeader title="Services" />
      <div className="flex flex-1 flex-col">
        <div className="@container/main flex flex-1 flex-col gap-2">
          <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
            <ServicesTable />
          </div>
        </div>
      </div>
    </>
  );
}
