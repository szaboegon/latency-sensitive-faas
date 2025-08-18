import { useMutation } from "@tanstack/react-query";
import DeploymentService from "../services/DeploymentService";
import type { RoutingTable } from "../models/models";

export const useModifyRoutingTable = () =>
  useMutation({
    mutationFn: (vars: { deploymentId: string; routingTable: RoutingTable }) =>
      DeploymentService.modifyRoutingTable(vars.deploymentId, vars.routingTable),
  });