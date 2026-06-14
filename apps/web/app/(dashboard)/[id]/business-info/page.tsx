import { SiteHeader } from "@repo/ui/components/ui/site-header";
import { FaqList } from "@/features/faq/components/faq-list";

export default function BusinessInfoPage() {
  return (
    <>
      <SiteHeader title="Business Info" />
      <div className="flex flex-1 flex-col">
        <div className="@container/main flex flex-1 flex-col gap-2">
          <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
            <div className="px-4 lg:px-6">
              <FaqList />
            </div>
          </div>
        </div>
      </div>
    </>
  );
}