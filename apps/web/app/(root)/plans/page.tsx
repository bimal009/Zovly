import { getPlans } from "@/features/plans/api/plans";
import Plans from "@/features/plans/components/Plans";
import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from "@tanstack/react-query";
import React from "react";

const page = async () => {
  const queryClient = new QueryClient();

  await queryClient.prefetchQuery({
    queryKey: ["posts"],
    queryFn: getPlans,
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <Plans />
    </HydrationBoundary>
  );
};

export default page;
