import { api } from "./apiClient";
import { PodItem } from "../models/k8sResource";

export const podApi = {
  // GET /api/v1/pods?namespace=production
  getPods: async (namespace: string) => {
    const response = await api.get<PodItem[]>("/pods", {
      params: { namespace },
    });
    return response.data;
  },

  // GET /api/v1/pods/:namespace/:pod/logs?container=nginx
  getPodLogs: async (namespace: string, pod: string, container: string) => {
    const response = await api.get<{ logs: string }>(
      `/pods/${namespace}/${pod}/logs`,
      { params: { container } }
    );
    return response.data.logs;
  },

  // Download File / Folder endpoint
  downloadContainerFile: async (namespace: string, pod: string, container: string, path: string) => {
    const response = await api.get(
      `/pods/${namespace}/${pod}/containers/${container}/download`,
      {
        params: { path },
        responseType: "blob",
      }
    );
    return response.data;
  },
};