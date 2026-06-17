import { SiteHeader } from "@repo/ui/components/ui/site-header";
import { InstagramConnection } from "@/features/connections/components/instagram-connection";

export default function InstagramConnectionPage() {
  return (
    <>
      <SiteHeader title="Instagram" />
      <div className="flex flex-1 flex-col px-4 py-4 md:px-6">
        <InstagramConnection />
      </div>
    </>
  );
}
