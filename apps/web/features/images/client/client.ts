import { useQuery } from "@tanstack/react-query";
import { fetchIKAuth } from "../api/image";

export const useGetIKAuth = () => {
  return useQuery({
    queryKey: ["auth-images"],
    queryFn: fetchIKAuth,
  });
};
