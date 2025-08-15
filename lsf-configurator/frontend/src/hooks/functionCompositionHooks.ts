import { useMutation } from "@tanstack/react-query";
import FunctionCompositionService from "../services/FunctionCompositionService";
import type { FunctionComposition, RoutingTable } from "../models/models";

export const useDeleteFunctionComposition = () =>
  useMutation({
    mutationFn: FunctionCompositionService.deleteFunctionComposition,
  });

export const useModifyRoutingTable = () =>
  useMutation({
    mutationFn: (vars: { functionCompositionId: string; routingTable: RoutingTable }) =>
      FunctionCompositionService.modifyRoutingTable(vars.functionCompositionId, vars.routingTable),
  });

export const useCreateFunctionComposition = () =>
  useMutation({
    mutationFn: (vars: { appId: string; functionComposition: FunctionComposition }) =>
      FunctionCompositionService.createFunctionComposition(vars.appId, vars.functionComposition),
  });
