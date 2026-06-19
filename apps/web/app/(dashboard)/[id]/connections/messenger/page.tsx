import { SiteHeader } from "@repo/ui/components/ui/site-header";
import { MessengerConnection } from "@/features/connections/components/messenger-connection";

export default function MessengerConnectionPage() {
  return (
    <>
      <SiteHeader title="Messenger" />
      <div className="flex flex-1 flex-col px-4 py-4 md:px-6">
        <MessengerConnection />
      </div>
    </>
  );
}
