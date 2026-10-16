import { api } from "./apiClient";

export interface K8sResourceQuery {
  namespace?: string;
  search?: string;
}

export const k8sApi = {
  // ==========================================
  // 1. Workloads (Deployments, StatefulSets, DaemonSets, CronJobs)
  // ==========================================
  getResources: async (resourceType: string, namespace?: string) => {
    // E.g., /deployments, /statefulsets, /services, /configmaps
    const response = await api.get<any[]>(`/${resourceType}`, {
      params: { namespace: namespace === "all" ? undefined : namespace },
    });
    return response.data;
  },

  getResourceYaml: async (
    resourceType: string,
    namespace: string,
    name: string,
  ) => {
    const response = await api.get<{ yaml: string }>(
      `/${resourceType}/${namespace}/${name}/yaml`,
    );
    return response.data.yaml;
  },

  updateResourceYaml: async (
    resourceType: string,
    namespace: string,
    name: string,
    yamlContent: string,
  ) => {
    const response = await api.put(
      `/${resourceType}/${namespace}/${name}/yaml`,
      { yaml: yamlContent },
    );
    return response.data;
  },

  deleteResource: async (
    resourceType: string,
    namespace: string,
    name: string,
  ) => {
    const response = await api.delete(`/${resourceType}/${namespace}/${name}`);
    return response.data;
  },

  // ==========================================
  // 2. Cluster Admin / Global Resources (Nodes, Namespaces - no namespace required)
  // ==========================================
  getClusterResources: async (resourceType: string) => {
    const response = await api.get<any[]>(`/${resourceType}`);
    return response.data;
  },

  // ==========================================
  // 3. Scaling & Operational Actions (Deployments / StatefulSets)
  // ==========================================
  scaleResource: async (
    resourceType: string,
    namespace: string,
    name: string,
    replicas: number,
  ) => {
    const response = await api.post(
      `/${resourceType}/${namespace}/${name}/scale`,
      { replicas },
    );
    return response.data;
  },

  restartResource: async (
    resourceType: string,
    namespace: string,
    name: string,
  ) => {
    const response = await api.post(
      `/${resourceType}/${namespace}/${name}/restart`,
    );
    return response.data;
  },
};
