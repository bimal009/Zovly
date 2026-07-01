import { SiteHeader } from "@repo/ui/components/ui/site-header";
import { FacebookConnection } from "@/features/connections/components/facebook-connection";

interface PageProps {
 params: Promise<{ id: string }>;
}

export default async function FacebookConnectionPage({ params }: PageProps) {
  const { id } = await params;
  return (
    <>
      <SiteHeader title="Facebook" />
      <div className="flex flex-1 flex-col px-4 py-4 md:px-6">
        <FacebookConnection businessId={id} />
      </div>
    </>
  );
}
