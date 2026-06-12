"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createService,
  deleteService,
  getServices,
  updateService,
  type CreateServiceInput,
  type ListServicesParams,
  type UpdateServiceInput,
} from "../api/services";

export const SERVICES_KEY = ["services"] as const;

export const useGetServices = (params?: ListServicesParams) => {
  return useQuery({
    queryKey: [...SERVICES_KEY, params],
    queryFn: () => getServices(params),
  });
};

export const useCreateService = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateServiceInput) => createService(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: SERVICES_KEY }),
  });
};

export const useUpdateService = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateServiceInput }) =>
      updateService(id, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: SERVICES_KEY }),
  });
};

export const useDeleteService = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteService(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: SERVICES_KEY }),
  });
};
