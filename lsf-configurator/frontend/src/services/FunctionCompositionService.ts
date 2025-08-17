import axios from "axios";
import paths from "../helpers/paths";
import type { FunctionComposition, RoutingTable } from "../models/models";

const FunctionCompositionService = {

  async deleteFunctionComposition(id: string): Promise<void> {
    await axios.delete(`${paths.functionCompositions}/${id}`);
  },

  async modifyRoutingTable(functionCompositionId: string, routingTable: RoutingTable): Promise<void> {
    await axios.put(`${paths.functionCompositions}/${functionCompositionId}/routing-table`, routingTable);
  },

  async createFunctionComposition(appId: string, fc: FunctionComposition, autoDeploy: boolean): Promise<void> {
    await axios.post(`${paths.functionCompositions}?app_id=${appId}&auto_deploy=${autoDeploy}`, fc);
  }
};

export default FunctionCompositionService;