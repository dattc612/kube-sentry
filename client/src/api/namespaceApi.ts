import { api } from "./apiClient";
import { NamespacesResponse } from "../models/k8sResource";
export const namespaceApi = {
  getNamespaces: async () => {
    const response = await api.get<NamespacesResponse>("/namespaces");
    return response.data;
  },
};
