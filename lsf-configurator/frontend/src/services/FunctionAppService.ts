import axios from "axios";
import type { FunctionApp } from "../models/models";

const FunctionAppService = {
  async fetchFunctionApps(): Promise<FunctionApp[]> {
    const response = await axios.get("/api/function-apps"); // Adjust the endpoint as needed
    return response.data;
  },

  async createFunctionApp(newApp: FunctionApp): Promise<FunctionApp> {
    const response = await axios.post("/api/function-apps", newApp);
    return response.data;
  },
};

export default FunctionAppService;
