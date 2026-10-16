import { useState, useEffect, useCallback } from "react";
import { podApi } from "../api/podApi";
import { PodItem } from "../models/k8sResource";

export function usePods(namespaceFilter: string) {
  const [pods, setPods] = useState<PodItem[]>([]);
  const [selectedPod, setSelectedPod] = useState<PodItem | null>(null);
  const [logs, setLogs] = useState<string>("");

  const [loading, setLoading] = useState<boolean>(false);
  const [logsLoading, setLogsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch pods list based on current namespace
  const fetchPods = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const ns = namespaceFilter === "all" ? "" : namespaceFilter;
      const data = await podApi.getPods(ns);

      setPods(data);
      if (data && data.length > 0) {
        setSelectedPod(data[0]);
      } else {
        setSelectedPod(null);
      }
    } catch (err: any) {
      setError(err.message || "Failed to fetch pods");
    } finally {
      setLoading(false);
    }
  }, [namespaceFilter]);

  // Fetch container logs
  const fetchLogs = useCallback(
    async (namespace: string, podName: string, containerName: string) => {
      try {
        setLogsLoading(true);
        const logContent = await podApi.getPodLogs(
          namespace,
          podName,
          containerName,
        );
        setLogs(logContent);
      } catch (err: any) {
        setLogs(`Failed to fetch logs: ${err.message}`);
      } finally {
        setLogsLoading(false);
      }
    },
    [],
  );

  useEffect(() => {
    fetchPods();
  }, [fetchPods]);

  return {
    pods,
    selectedPod,
    setSelectedPod,
    logs,
    loading,
    logsLoading,
    error,
    fetchPods,
    fetchLogs,
  };
}
