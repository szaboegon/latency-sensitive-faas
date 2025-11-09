import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import FunctionAppService from "../services/FunctionAppService";
import { functionAppsMock } from "../mocks/functionAppsMock";
import type { FunctionApp } from "../models/models";
import type { BulkCreateRequest, FunctionAppCreateDto } from "../models/dto";

const useMockData = import.meta.env.VITE_USE_MOCK_DATA === "true";

export const useFunctionApps = () => {
  return useQuery<FunctionApp[]>({
    queryKey: ["functionApps"],
    queryFn: useMockData
      ? async () => functionAppsMock // Return mock data if toggled
      : FunctionAppService.fetchFunctionApps,
  });
};

export const useFunctionAppById = (id: string) => {
  return useQuery<FunctionApp | null>({
    queryKey: ["functionApps", id],
    queryFn: useMockData
      ? async () => functionAppsMock.find((app) => app.id === id) || null
      : () => FunctionAppService.fetchFunctionAppById(id),
  });
};

export const useCreateFunctionApp = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (vars: {
      functionApp: FunctionAppCreateDto;
      files: FileList;
    }) => FunctionAppService.createFunctionApp(vars.functionApp, vars.files),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["functionApps"] });
    },
    meta: {
      successMessage: "Function app created successfully!",
    },
  });
};

export const useBulkCreateFunctionApp = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (vars: { req: BulkCreateRequest; files: FileList }) =>
      FunctionAppService.bulkCreateFunctionApp(vars.req, vars.files),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["functionApps"] });
    },
    meta: {
      successMessage: "Function apps created successfully!",
    },
  });
};

export const useDeleteFunctionApp = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: FunctionAppService.deleteFunctionApp,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["functionApps"] });
    },
    meta: {
      successMessage: "Function app deleted successfully!",
    },
  });
};

export const useUpdateFunctionAppLatencyLimit = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (vars: { id: string; latencyLimit: number }) =>
      FunctionAppService.updateFunctionAppLatencyLimit(
        vars.id,
        vars.latencyLimit
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["functionApps"] });
    },
    meta: {
      successMessage: "Function app latency limit updated successfully!",
    },
  });
};
