import axiosInstance from "./axios";
import type { FunctionApp } from "../models/models";
import paths from "../helpers/paths";
import type { BulkCreateRequest, FunctionAppCreateDto } from "../models/dto";
import { keysToSnakeCase } from "../helpers/utilities";

const FunctionAppService = {
  async fetchFunctionApps(): Promise<FunctionApp[]> {
    const response = await axiosInstance.get(`${paths.functionApps}/`);
    return response.data;
  },

  async fetchFunctionAppById(id: string): Promise<FunctionApp | null> {
    const response = await axiosInstance.get(`${paths.functionApps}/${id}`);
    return response.data;
  },

  async createFunctionApp(newApp: FunctionAppCreateDto, files: FileList): Promise<void> {
    const formData = new FormData();
    formData.append("json", JSON.stringify(keysToSnakeCase(newApp)));
    Array.from(files).forEach((file) => {
      formData.append("files", file);
    });

    await axiosInstance.post(`${paths.functionApps}/`, formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
  },

  async bulkCreateFunctionApp(req: BulkCreateRequest, files: FileList): Promise<void> {
    const formData = new FormData();
    formData.append("json", JSON.stringify(keysToSnakeCase(req)));
    Array.from(files).forEach((file) => {
      formData.append("files", file);
    });

    await axiosInstance.post(`${paths.functionApps}/bulk`, formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
  },

  async deleteFunctionApp(id: string): Promise<void> {
    await axiosInstance.delete(`${paths.functionApps}/${id}`);
  }
};

export default FunctionAppService;
