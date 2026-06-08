import { getPlans } from "@/features/plans/api/plans";
import { CheckoutClient } from "@/features/plans/components/CheckoutClient";
import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from "@tanstack/react-query";
import { Suspense } from "react";

const CheckoutPage = async () => {
  const queryClient = new QueryClient();

  await queryClient.prefetchQuery({
    queryKey: ["plans"],
    queryFn: getPlans,
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <Suspense>
        <CheckoutClient />
      </Suspense>
    </HydrationBoundary>
  );
};

export default CheckoutPage;
