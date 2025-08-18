import { useMutation } from "@tanstack/react-query";
import FunctionCompositionService from "../services/FunctionCompositionService";
import type { FunctionComposition } from "../models/models";

export const useDeleteFunctionComposition = () =>
  useMutation({
    mutationFn: FunctionCompositionService.deleteFunctionComposition,
  });

export const useCreateFunctionComposition = () =>
  useMutation({
    mutationFn: (vars: { appId: string; functionComposition: FunctionComposition }) =>
      FunctionCompositionService.createFunctionComposition(vars.appId, vars.functionComposition),
  });
