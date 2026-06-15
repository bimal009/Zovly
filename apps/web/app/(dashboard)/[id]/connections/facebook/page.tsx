import { SiteHeader } from "@repo/ui/components/ui/site-header";
import { FacebookConnection } from "@/features/connections/components/facebook-connection";

export default function FacebookConnectionPage() {
  return (
    <>
      <SiteHeader title="Facebook" />
      <div className="flex flex-1 flex-col px-4 py-4 md:px-6">
        <FacebookConnection />
      </div>
    </>
  );
}
