import { FacebookMessagingEvent, FacebookWebhookChange, FacebookWebhookPayload, isDeliveryEvent, isMessageEvent, isOptinEvent, isPostbackEvent, isReadEvent, isReferralEvent } from "@repo/types";
import { ValidationError } from "../../lib/errors";





export function facebookWebhookChallenge(
  mode: unknown,
  token: unknown,
  challenge: unknown
): string {
  if (mode === "subscribe" && token === process.env.FACEBOOK_WEBHOOK_VERIFY_TOKEN) {
    return String(challenge);
  }
  throw new ValidationError("Facebook webhook verification failed.");
}

export async function processFacebookWebhookPayload(body: FacebookWebhookPayload): Promise<void> {
  if (body?.object !== "page") return;

  for (const entry of body.entry ?? []) {
    const pageId = entry.id;

    for (const change of entry.changes ?? []) {
      await dispatchFacebookEvent(pageId, change);
    }

    for (const messagingEvent of entry.messaging ?? []) {
      await dispatchFacebookMessagingEvent(pageId,  messagingEvent);
    }
  }
}

async function dispatchFacebookEvent(pageId: string, change: FacebookWebhookChange) {
  switch (change.field) {
    case "feed": {
      const value = change.value;
      if (value.item === "comment" && value.verb === "add") {
        console.log(`New comment on Page ${pageId}:`, value);
        // TODO: comment automation hook
      }
      break;
    }

    case "mention":
      console.log(`Page ${pageId} mentioned:`, change.value);
      // TODO: mention automation hook
      break;

    case "leadgen":
      console.log(`New lead on Page ${pageId}:`, change.value);
      // TODO: fetch full lead via change.value.leadgen_id, trigger automation
      break;

    case "ratings":
      console.log(`New rating on Page ${pageId}:`, change.value);
      // TODO: review automation hook
      break;

    default:
      console.log(`Unhandled Facebook field "${change.field}" for Page ${pageId}:`, change.value);
  }
}

async function dispatchFacebookMessagingEvent(pageId: string, event: FacebookMessagingEvent) {

  if (isMessageEvent(event)) {
    const isEcho = event.message.is_echo === true
    const userPsid = isEcho ? event.recipient.id : event.sender.id

    if (isEcho) {
      // outbound message your business sent (via Send API or manual inbox reply)
      console.log(`Outbound (echo) on Page ${pageId} to ${userPsid}:`, event.message);
      // TODO: log via messageService.create(..., direction: "out")
    } else {
      // inbound message from the customer
      console.log(`Inbound message on Page ${pageId} from ${userPsid}:`, event.message);
      // TODO: log via messageService.create(..., direction: "in") + automation hook
    }
  } else if (isDeliveryEvent(event)) {
    const userPsid = event.recipient.id
    // TODO: mark delivered — event.delivery.mids, event.delivery.watermark
  } else if (isReadEvent(event)) {
    const userPsid = event.recipient.id
    // TODO: mark read/seen — event.read.watermark
  } else if (isPostbackEvent(event)) {
    const userPsid = event.sender.id
    // TODO: button click automation — event.postback.payload
  } else if (isReferralEvent(event) || isOptinEvent(event)) {
    const userPsid = event.sender.id
    // TODO: m.me referral / opt-in automation
  }
}
