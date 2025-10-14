import { useQuery } from "@tanstack/react-query";
import ResultsService from "../services/ResultsService";
import { resultsMock } from "../mocks/resultsMock";
import type { AppResult } from "../models/models";

const useMockData = import.meta.env.VITE_USE_MOCK_DATA === "true";

export const useGetAppResults = (appId: string, count: number) => {
  return useQuery<AppResult[]>({
    queryKey: ["appResults", appId, count],
    queryFn: async () => {
      if (useMockData) {
        return resultsMock;
      }
      const response = await ResultsService.getAppResults(appId, count);
      return response;
    },
    enabled: !!appId,
  });
};
