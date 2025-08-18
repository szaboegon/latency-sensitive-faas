import { useMutation } from "@tanstack/react-query";
import DeploymentService from "../services/DeploymentService";
import type { Deployment, RoutingTable } from "../models/models";

export const useCreateDeployment = () =>
  useMutation({
    mutationFn: (deployment: Deployment) => DeploymentService.createDeployment(deployment),
  });

export const useModifyRoutingTable = () =>
  useMutation({
    mutationFn: (vars: { deploymentId: string; routingTable: RoutingTable }) =>
      DeploymentService.modifyRoutingTable(vars.deploymentId, vars.routingTable),
  });