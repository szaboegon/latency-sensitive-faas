import axios from "axios";
import paths from "../helpers/paths";
import type { RoutingTable } from "../models/models";

const FunctionCompositionService = {

  async deleteFunctionComposition(id: string): Promise<void> {
    await axios.delete(`${paths.functionCompositions}/${id}`);
  },

  async modifyRoutingTable(functionCompositionId: string, routingTable: RoutingTable): Promise<void> {
    await axios.put(`${paths.functionCompositions}/${functionCompositionId}/routing-table`, routingTable);
  }
};

export default FunctionCompositionService;