import { randomUUID } from "crypto";
import { and, eq } from "drizzle-orm";
import { db } from "../../config/db/db";
import { verification } from "../../config/db/schema/user";
import { appCredentials } from "../../config/db/schema/schema";
import { decryptToken, encryptToken } from "../../lib/encrypt";
import {
  BadRequestError,
  NotFoundError,
  ServiceUnavailableError,
  ValidationError,
} from "../../lib/errors";

const GRAPH_VERSION = "v25.0";
const GRAPH_BASE = `https://graph.facebook.com/${GRAPH_VERSION}`;

const metaAppId = process.env.META_APP_ID!;
const metaAppSecret = process.env.META_APP_SECRET!;
const metaRedirectUri = process.env.META_REDIRECT_URI!;

const FACEBOOK_SCOPES = [
  "pages_show_list",
  "pages_manage_posts",
  "pages_read_engagement",
  "pages_manage_metadata",
  "pages_read_user_content",
  "pages_messaging",
  "business_management",
];

const WEBHOOK_SUBSCRIBED_FIELDS = [
  "feed", // comments, posts, likes on the Page — comments arrive here as item:"comment"
  "mention", // Page mentioned in someone else's post
  "leadgen", // lead form submissions

  "messages", // inbound Messenger messages — core for automation/bots
  "messaging_postbacks", // button/quick-reply clicks — needed for bot flows
  "messaging_optins", // "Send to Messenger" plugin opt-ins
  "message_deliveries", // delivered receipts
  "message_reads", // "seen" receipts
  "message_reactions", // emoji reactions on messages
  "message_echoes", // messages YOUR page/bot sent — lets you track your own automated replies
  "messaging_referrals", // m.me links / ref params — useful as automation entry points
  "messaging_handovers", // only needed if you'll run multiple bots/agents on the same Page
  "standby", // only needed alongside messaging_handovers
];

const STATE_TTL_MS = 10 * 60 * 1000;
const STATE_PREFIX = "fb-connect:";

// Meta Graph API error codes worth special-casing
const META_ERROR_APP_NOT_INSTALLED = 100;

interface FacebookPage {
  id: string;
  name: string;
  access_token: string;
  picture?: { data?: { url?: string } };
}

interface StateValue {
  businessId: string;
}

interface MetaErrorInfo {
  code?: number;
  message: string;
  raw: unknown;
}

async function parseMetaError(res: Response): Promise<MetaErrorInfo> {
  const body = await res.json().catch(() => null);
  return {
    code: body?.error?.code,
    message: body?.error?.message ?? res.statusText,
    raw: body,
  };
}

export async function buildFacebookConnectUrl(businessId: string): Promise<string> {
  if (!businessId) {
    throw new ValidationError("businessId is required to start Facebook connect flow.");
  }

  const state = randomUUID();

  await db.insert(verification).values({
    id: randomUUID(),
    identifier: STATE_PREFIX + state,
    value: JSON.stringify({ businessId } satisfies StateValue),
    expiresAt: new Date(Date.now() + STATE_TTL_MS),
  });

  const params = new URLSearchParams({
    client_id: metaAppId,
    redirect_uri: metaRedirectUri,
    state,
    scope: FACEBOOK_SCOPES.join(","),
    response_type: "code",
  });

  return `https://www.facebook.com/${GRAPH_VERSION}/dialog/oauth?${params.toString()}`;
}

async function consumeState(state: string): Promise<StateValue> {
  if (!state) {
    throw new ValidationError("Missing OAuth state parameter.");
  }

  const identifier = STATE_PREFIX + state;

  const row = await db.query.verification.findFirst({
    where: (t, { eq }) => eq(t.identifier, identifier),
  });

  if (!row) {
    throw new NotFoundError("Invalid or already-used Facebook OAuth state.");
  }

  if (row.expiresAt < new Date()) {
    await db.delete(verification).where(eq(verification.id, row.id));
    throw new BadRequestError("Facebook OAuth state expired. Please try connecting again.");
  }

  // Consume it immediately so it can't be replayed
  await db.delete(verification).where(eq(verification.id, row.id));

  return JSON.parse(row.value) as StateValue;
}

export async function handleFacebookCallback(state: string, code: string) {
  if (!code) {
    throw new ValidationError("Missing authorization code from Facebook.");
  }

  const { businessId } = await consumeState(state);

  const tokenRes = await fetch(
    `${GRAPH_BASE}/oauth/access_token?` +
      new URLSearchParams({
        client_id: metaAppId,
        client_secret: metaAppSecret,
        redirect_uri: metaRedirectUri,
        code,
      })
  );
  if (!tokenRes.ok) {
    const { code: errCode, message } = await parseMetaError(tokenRes);
    throw new ServiceUnavailableError(`Facebook token exchange failed (code ${errCode}): ${message}`);
  }
  const { access_token } = await tokenRes.json();

  const longLivedRes = await fetch(
    `${GRAPH_BASE}/oauth/access_token?` +
      new URLSearchParams({
        grant_type: "fb_exchange_token",
        client_id: metaAppId,
        client_secret: metaAppSecret,
        fb_exchange_token: access_token,
      })
  );
  if (!longLivedRes.ok) {
    const { code: errCode, message } = await parseMetaError(longLivedRes);
    throw new ServiceUnavailableError(
      `Facebook long-lived token exchange failed (code ${errCode}): ${message}`
    );
  }
  const longLived: { access_token: string; expires_in?: number } = await longLivedRes.json();

  const pagesRes = await fetch(
    `${GRAPH_BASE}/me/accounts?` +
      new URLSearchParams({
        access_token: longLived.access_token,
        fields: "id,name,access_token,picture{url}",
      })
  );
  if (!pagesRes.ok) {
    const { code: errCode, message } = await parseMetaError(pagesRes);
    throw new ServiceUnavailableError(`Failed to fetch Facebook Pages (code ${errCode}): ${message}`);
  }
  const { data: pages }: { data: FacebookPage[] } = await pagesRes.json();

  if (!pages?.length) {
    throw new NotFoundError("No Facebook Pages found. User must be an admin on at least one Page.");
  }

  const result: { id: string; name: string; image?: string }[] = [];

  for (const page of pages) {
    const image = page.picture?.data?.url;

    await db
      .insert(appCredentials)
      .values({
        businessId,
        appName: "facebook",
        accessToken: encryptToken(page.access_token),
        tokenExpiresAt: null,
        scopes: FACEBOOK_SCOPES,
        platformAccountId: page.id,
        platformAccountName: page.name,
        platformAccountImage: image ?? null,
        isActive: false,
        connectedAt: new Date(),
      })
      .onConflictDoUpdate({
        target: [appCredentials.businessId, appCredentials.appName, appCredentials.platformAccountId],
        set: {
          accessToken: encryptToken(page.access_token),
          tokenExpiresAt: null,
          platformAccountName: page.name,
          platformAccountImage: image ?? null,
          connectedAt: new Date(),
          errorMessage: null,
          updatedAt: new Date(),
        },
      });

    result.push({ id: page.id, name: page.name, image });
  }

  return result;
}

async function subscribePageToWebhooks(pageId: string, pageAccessToken: string): Promise<void> {
  const res = await fetch(
    `${GRAPH_BASE}/${pageId}/subscribed_apps?` +
      new URLSearchParams({
        subscribed_fields: WEBHOOK_SUBSCRIBED_FIELDS.join(","),
        access_token: pageAccessToken,
      }),
    { method: "POST" }
  );

  if (!res.ok) {
    const { code, message } = await parseMetaError(res);
    throw new ServiceUnavailableError(`Failed to subscribe Page to webhooks (code ${code}): ${message}`);
  }
}

async function unsubscribePageFromWebhooks(pageId: string, pageAccessToken: string): Promise<void> {
  const res = await fetch(
    `${GRAPH_BASE}/${pageId}/subscribed_apps?` +
      new URLSearchParams({ access_token: pageAccessToken }),
    { method: "DELETE" }
  );

  if (!res.ok) {
    const { code, message } = await parseMetaError(res);

    if (code === META_ERROR_APP_NOT_INSTALLED) {
      return;
    }

    throw new ServiceUnavailableError(`Failed to unsubscribe Page from webhooks (code ${code}): ${message}`);
  }
}

export async function listFacebookPages(businessId: string) {
  const rows = await db.query.appCredentials.findMany({
    where: (t, { eq, and }) => and(eq(t.businessId, businessId), eq(t.appName, "facebook")),
  });

  return {
    connected: rows.length > 0,
    pages: rows.map((row) => ({
      id: row.id,
      platform_account_id: row.platformAccountId,
      platform_account_name: row.platformAccountName,
      platform_account_image: row.platformAccountImage,
      is_active: row.isActive,
      connected_at: row.connectedAt,
      error_message: row.errorMessage,
    })),
  };
}

export async function toggleFacebookPage(
  businessId: string,
  platformAccountId: string,
  nextIsActive: boolean
) {
  const row = await db.query.appCredentials.findFirst({
    where: (t, { eq, and }) =>
      and(
        eq(t.businessId, businessId),
        eq(t.appName, "facebook"),
        eq(t.platformAccountId, platformAccountId)
      ),
  });

  if (!row) {
    throw new NotFoundError("No Facebook Page connection found for this business.");
  }
  if (!row.accessToken) {
    throw new BadRequestError("This Page has no stored access token. Reconnect Facebook first.");
  }

  const pageAccessToken = decryptToken(row.accessToken);

  try {
    if (nextIsActive) {
      await subscribePageToWebhooks(platformAccountId, pageAccessToken);
    } else {
      await unsubscribePageFromWebhooks(platformAccountId, pageAccessToken);
    }
  } catch (err) {
    await db
      .update(appCredentials)
      .set({ errorMessage: (err as Error).message, updatedAt: new Date() })
      .where(eq(appCredentials.id, row.id));
    throw err;
  }

  const [updated] = await db
    .update(appCredentials)
    .set({
      isActive: nextIsActive,
      webhookSubscribedAt: nextIsActive ? new Date() : null,
      disconnectedAt: nextIsActive ? null : new Date(),
      errorMessage: null,
      updatedAt: new Date(),
    })
    .where(eq(appCredentials.id, row.id))
    .returning();

  return updated.id;
}