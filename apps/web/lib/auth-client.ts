import { createAuthClient } from "better-auth/react";
import { inferAdditionalFields } from "better-auth/client/plugins";

export const authClient = createAuthClient({
  baseURL: process.env.NEXT_PUBLIC_BETTER_AUTH_URL!,
  basePath: "/api/v1/auth",

  plugins: [
    inferAdditionalFields({
      user: {
        role: { type: "string" },
        isOnboarded: { type: "boolean" },
      },
    }),
  ],
});

export const { signIn, signUp, signOut, useSession } = authClient;
