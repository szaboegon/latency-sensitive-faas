import axiosInstance from "./axios";
import paths from "../helpers/paths";
import type { FunctionComposition } from "../models/models";

const FunctionCompositionService = {
  async deleteFunctionComposition(id: string): Promise<void> {
    await axiosInstance.delete(`${paths.functionCompositions}/${id}`);
  },
  async createFunctionComposition(appId: string, fc: FunctionComposition): Promise<void> {
    await axiosInstance.post(`${paths.functionCompositions}?app_id=${appId}`, fc);
  },
};

export default FunctionCompositionService;