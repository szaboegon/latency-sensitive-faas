import axios from "axios";
import paths from "../helpers/paths";
import type { RoutingTable } from "../models/models";

const DeploymentService = {
  async modifyRoutingTable(deploymentId: string, routingTable: RoutingTable): Promise<void> {
    await axios.put(`${paths.deployments}/${deploymentId}/routing-table`, routingTable);
  },
}

export default DeploymentService;