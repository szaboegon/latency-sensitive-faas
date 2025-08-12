import { useMutation } from "@tanstack/react-query";
import FunctionCompositionService from "../services/FunctionCompositionService";
import type { RoutingTable } from "../models/models";

export const useDeleteFunctionComposition = () =>
  useMutation({
    mutationFn: FunctionCompositionService.deleteFunctionComposition,
  });

export const useModifyRoutingTable = () =>
  useMutation({
    mutationFn: (vars: { functionCompositionId: number; routingTable: RoutingTable }) =>
      FunctionCompositionService.modifyRoutingTable(vars.functionCompositionId, vars.routingTable),
  });