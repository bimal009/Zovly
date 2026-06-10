import { betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";
import { db } from "@repo/database";
import * as schema from "@repo/database/schema";
import { redis } from "./redis";

export const auth = betterAuth({
  database: drizzleAdapter(db, {
    provider: "pg",
    schema: {
      ...schema,
      verification: schema.verification,
      user: schema.user,
    },
  }),
  secret: process.env.BETTER_AUTH_SECRET! as string,
  baseURL: process.env.BETTER_AUTH_URL! as string,

  // Redis secondary storage (manual get/set/delete)
  secondaryStorage: {
    get: async (key) => {
      return await redis.get(key); // returns the stored string, or null
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
    storage: "secondary-storage", // push rate-limit counters into Redis too
  },

  emailAndPassword: {
    enabled: true,
    requireEmailVerification: process.env.NODE_ENV === "production",
  },

  socialProviders: {
    google: {
      clientId: process.env.GOOGLE_CLIENT_ID as string,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET as string,
      redirectURI: `${process.env.BETTER_AUTH_URL || "http://localhost:3000"}/api/auth/callback/google`,
    },
  },

  user: {
    additionalFields: {
      role: {
        type: "string",
        required: false,
        defaultValue: "user",
        input: false,
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
    storeSessionInDatabase: true, // keep sessions in Postgres so your Go GetByToken still works
    cookieCache: {
      enabled: true,
      maxAge: 60 * 5,
    },
  },

  advanced: {
    crossSubDomainCookies: {
      enabled: true,
    },
    cookiePrefix: "zovly",
    useSecureCookies: process.env.NODE_ENV === "production",
    cookieSameSite: process.env.NODE_ENV === "production" ? "strict" : "lax",
  },
});

export type Auth = typeof auth;
export type Session = Auth["$Infer"]["Session"];
export type User = Session["user"];
