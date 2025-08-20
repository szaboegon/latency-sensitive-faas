import axiosInstance from "./axios";
import paths from "../helpers/paths";
import type { RoutingTable } from "../models/models";
import type { DeploymentCreateDto } from "../models/dto";

const DeploymentService = {
  async createDeployment(deployment: DeploymentCreateDto): Promise<void> {
    await axiosInstance.post(paths.deployments, deployment);
  },
  async modifyRoutingTable(deploymentId: string, routingTable: RoutingTable): Promise<void> {
    await axiosInstance.put(`${paths.deployments}/${deploymentId}/routing-table`, routingTable);
  },
};

export default DeploymentService;