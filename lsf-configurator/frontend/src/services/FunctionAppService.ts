import axiosInstance from "./axios";
import type { FunctionApp } from "../models/models";
import paths from "../helpers/paths";

const FunctionAppService = {
  async fetchFunctionApps(): Promise<FunctionApp[]> {
    const response = await axiosInstance.get(paths.apps);
    return response.data;
  },

  async fetchFunctionAppById(id: string): Promise<FunctionApp | null> {
    const response = await axiosInstance.get(`${paths.apps}/${id}`);
    return response.data;
  },

  async createFunctionApp(newApp: FunctionApp, files: FileList): Promise<FunctionApp> {
    const formData = new FormData();
    formData.append("json", JSON.stringify(newApp));
    Array.from(files).forEach((file) => {
      formData.append("files", file);
    });

    const response = await axiosInstance.post(`${paths.apps}`, formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
    return response.data;
  },

  async deleteFunctionApp(id: string): Promise<void> {
    await axiosInstance.delete(`${paths.apps}/${id}`);
  }
};

export default FunctionAppService;
