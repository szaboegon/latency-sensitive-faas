import { useMutation, useQueryClient } from "@tanstack/react-query";
import FunctionCompositionService from "../services/FunctionCompositionService";
import type { FunctionCompositionCreateDto } from "../models/dto";

export const useDeleteFunctionComposition = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: FunctionCompositionService.deleteFunctionComposition,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["functionApps"] }); 
    },
    meta: {
      successMessage: "Function composition deletion started, this might take a while...",
    },
  });
};

export const useCreateFunctionComposition = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (vars: { appId: string; functionComposition: FunctionCompositionCreateDto }) =>
      FunctionCompositionService.createFunctionComposition(vars.appId, vars.functionComposition),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["functionApps"] }); 
    },
    meta: {
      successMessage: "Function composition created successfully!",
    },
  });
};
