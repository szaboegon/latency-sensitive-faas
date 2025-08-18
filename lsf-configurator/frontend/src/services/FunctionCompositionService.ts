import axios from "axios";
import paths from "../helpers/paths";
import type { FunctionComposition } from "../models/models";

const FunctionCompositionService = {

  async deleteFunctionComposition(id: string): Promise<void> {
    await axios.delete(`${paths.functionCompositions}/${id}`);
  },

  async createFunctionComposition(appId: string, fc: FunctionComposition): Promise<void> {
    await axios.post(`${paths.functionCompositions}?app_id=${appId}`, fc);
  }
};

export default FunctionCompositionService;