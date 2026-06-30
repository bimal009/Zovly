import { betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";
import { db } from "../config/db/db";
import * as schema from "../config/db/schema/schema";
import { client as redis } from "../config/redis";
import "dotenv/config";
const requiredEnvVars = [
  "BETTER_AUTH_SECRET",
  "BETTER_AUTH_URL",
  "FRONTEND_URL",
  "GOOGLE_CLIENT_ID",
  "GOOGLE_CLIENT_SECRET",
] as const;

for (const key of requiredEnvVars) {
  if (!process.env[key]) {
    throw new Error(`Missing required env var: ${key}`);
  }
}

const isProd = process.env.NODE_ENV === "production";

export const auth = betterAuth({
  database: drizzleAdapter(db, {
    provider: "pg",
    schema: {
      ...schema,
      verification: schema.verification,
      user: schema.user,
    },
  }),

  secret: process.env.BETTER_AUTH_SECRET!,
  baseURL: process.env.BETTER_AUTH_URL!,
  basePath: "/api/v1/auth",
  trustedOrigins: [process.env.FRONTEND_URL!],

  secondaryStorage: {
    get: async (key) => {
      return await redis.get(key);
    },
    set: async (key, value, ttl) => {
      if (ttl) await redis.set(key, value, "EX", ttl);
      else await redis.set(key, value);
    },
    delete: async (key) => {
      await redis.del(key);
    },
  },

  rateLimit: {
    window: 60,
    max: 10,
    storage: "secondary-storage",
  },

  emailAndPassword: {
    enabled: true,
    requireEmailVerification: isProd,
  },

  emailVerification: {
    sendOnSignUp: isProd,
    autoSignInAfterVerification: true,
    sendVerificationEmail: async ({ user, url }) => {
      if (!isProd) {
        console.log(`[dev] verification email for ${user.email}: ${url}`);
        return;
      }
    },
  },

  socialProviders: {
    google: {
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
      callbackURL: `${process.env.BETTER_AUTH_URL}/api/v1/auth/callback/google`,
      errorCallbackURL: `${process.env.BETTER_AUTH_URL}/api/v1/auth/callback/google?error=true`,
    },
  },

  user: {
    additionalFields: {
      role: {
        type: "string",
        required: false,
        defaultValue: "user",
        input: true,
      },
      isOnboarded: {
        type: "boolean",
        required: false,
        defaultValue: false,
        input: true,
      },
    },
  },

  session: {
    expiresIn: 60 * 60 * 24 * 7,
    updateAge: 60 * 60 * 24,
    storeSessionInDatabase: true,
    cookieCache: {
      enabled: true,
      maxAge: 60 * 5,
    },
  },

  advanced: {
    crossSubDomainCookies: {
      enabled: isProd,
      domain: isProd ? process.env.COOKIE_DOMAIN : undefined,
    },
    cookiePrefix: "zovly",
    useSecureCookies: isProd,
    cookieSameSite: isProd ? "strict" : "lax",
  },
});

export type Auth = typeof auth;
export type Session = Auth["$Infer"]["Session"];
export type User = Session["user"];
