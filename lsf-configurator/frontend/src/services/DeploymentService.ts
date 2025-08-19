import axiosInstance from "./axios";
import paths from "../helpers/paths";
import type { Deployment, RoutingTable } from "../models/models";

const DeploymentService = {
  //TODO routing table is empty here :(
  async createDeployment(deployment: Deployment): Promise<void> {
    await axiosInstance.post(paths.deployments, deployment);
  },
  async modifyRoutingTable(deploymentId: string, routingTable: RoutingTable): Promise<void> {
    await axiosInstance.put(`${paths.deployments}/${deploymentId}/routing-table`, routingTable);
  },
};

export default DeploymentService;