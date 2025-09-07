import axiosInstance from "./axios";
import paths from "../helpers/paths";
import type { FunctionCompositionCreateDto } from "../models/dto";

const FunctionCompositionService = {
  async deleteFunctionComposition(id: string): Promise<void> {
    await axiosInstance.delete(`${paths.functionCompositions}/${id}`);
  },
  async createFunctionComposition(appId: string, fc: FunctionCompositionCreateDto): Promise<void> {
    fc.functionAppId = appId;
    await axiosInstance.post(`${paths.functionCompositions}`, fc);
  },
};

export default FunctionCompositionService;