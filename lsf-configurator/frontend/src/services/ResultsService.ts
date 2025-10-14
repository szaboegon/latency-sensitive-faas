import paths from "../helpers/paths";
import axiosInstance from "./axios";

const ResultsService = {
  async getAppResults(appId: string, count: number): Promise<string[]> {
    const response = await axiosInstance.get(`${paths.results}/${appId}`, {
      params: { count: count },
    });
    return response.data;
  },
};

export default ResultsService;
