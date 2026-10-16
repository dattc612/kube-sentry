import { useState, useEffect, useCallback } from "react";
import { k8sApi } from "../api/k8sApi";

// List of cluster-wide resources that don't belong to a namespace
const CLUSTER_WIDE_RESOURCES = ["nodes", "namespaces", "pvs"];

export const useK8sResources = (resourceType: string, namespace: string) => {
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const fetchResources = useCallback(async () => {
    if (!resourceType) return;

    setLoading(true);
    setError(null);

    try {
      let result;
      const isClusterWide = CLUSTER_WIDE_RESOURCES.includes(resourceType);

      if (isClusterWide) {
        result = await k8sApi.getClusterResources(resourceType);
      } else {
        result = await k8sApi.getResources(resourceType, namespace);
      }

      setData(result || []);
    } catch (err: any) {
      setError(err.message || "Failed to fetch cluster resources");
      setData([]);
    } finally {
      setLoading(false);
    }
  }, [resourceType, namespace]);

  useEffect(() => {
    fetchResources();
  }, [fetchResources]);

  return { data, loading, error, refetch: fetchResources };
};
