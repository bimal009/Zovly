import Plans from "@/features/plans/components/Plans";
import { auth } from "@/lib/auth";
import { headers } from "next/headers";
import React from "react";

const page = async () => {
  const session = await auth.api.getSession({ headers: await headers() });
  console.log(session?.session.token);
  return (
    <div>
      <Plans />
    </div>
  );
};

export default page;
