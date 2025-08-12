import { useQuery, useMutation } from "@tanstack/react-query";
import FunctionAppService from "../services/FunctionAppService";
import { functionAppsMock } from "../mocks/functionAppsMock";
import type { FunctionApp } from "../models/models";

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
    queryKey: ["functionApp", id],
    queryFn: useMockData
      ? async () => functionAppsMock.find((app) => app.id === id) || null
      : () => FunctionAppService.fetchFunctionAppById(id),
  });
};

export const useCreateFunctionApp = () =>
  useMutation({
    mutationFn: (vars: { functionApp: FunctionApp; files: FileList }) =>
      FunctionAppService.createFunctionApp(vars.functionApp, vars.files)
  });

export const useDeleteFunctionApp = () =>
  useMutation({
    mutationFn: FunctionAppService.deleteFunctionApp,
  });
