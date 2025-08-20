import { useMutation, useQueryClient } from "@tanstack/react-query";
import DeploymentService from "../services/DeploymentService";
import type {  RoutingTable } from "../models/models";
import type {  DeploymentCreateDto } from "../models/dto";

export const useCreateDeployment = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (deployment: DeploymentCreateDto) => DeploymentService.createDeployment(deployment),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["functionApps"] });
    },
    meta: {
      successMessage: "Deployment created successfully!",
    },
  });
};

export const useModifyRoutingTable = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (vars: { deploymentId: string; routingTable: RoutingTable }) =>
      DeploymentService.modifyRoutingTable(vars.deploymentId, vars.routingTable),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["functionApps"] });
    },
    meta: {
      successMessage: "Routing table updated successfully!",
    },
  });
};