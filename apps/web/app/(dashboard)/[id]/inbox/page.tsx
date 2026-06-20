import { SiteHeader } from "@repo/ui/components/ui/site-header";
import { InboxView } from "@/features/inbox/components/inbox-view";

export default function InboxPage() {
  return (
    <>
      <SiteHeader title="Inbox" />
      <InboxView />
    </>
  );
}
