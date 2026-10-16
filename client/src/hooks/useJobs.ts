import { useCallback, useEffect, useState } from "react";
import { JobItem } from "../models/jobs";
import { jobsApi } from "../api/jobApi";

interface UseJobParams {
  namespace: string;
  pod: string;
  container: string;
}

export function useJobs({ namespace, pod, container }: UseJobParams) {
  const [running, setRunning] = useState(false);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [status, setStatus] = useState<
    "idle" | "running" | "success" | "error"
  >("idle");
  const [jobs, setJobs] = useState<JobItem[]>([]);

  const [output, setOutput] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleRunJob = async (id: number) => {
    setRunning(true);
    setStatus("running");
    setError(null);
    setOutput(null);

    try {
      const res = await jobsApi.runJob(id, namespace, pod, container);
      setOutput(res.output || null);
      setStatus("success");
    } catch (err: any) {
      setStatus("error");

      const errData = err.response?.data;

      setError(errData?.error || err.message || "Failed to run the job");

      if (errData?.output) {
        setOutput(errData.output);
      }
    } finally {
      setRunning(false);
    }
  };

  const loadAvalableJobs = useCallback(async () => {
    try {
      setLoadingJobs(true);
      setError(null);
      const data = await jobsApi.listJobs(namespace, pod, container);
      setJobs(data);
    } catch (err: any) {
      setError(err.message || "Failed to fetch job list");
    } finally {
      setLoadingJobs(false);
    }
  }, [pod, namespace]);

  useEffect(() => {
    loadAvalableJobs();
  }, [loadAvalableJobs]);

  return {
    jobs,
    status,
    setJobs,
    loadingJobs,
    running,
    error,
    output,
    handleRunJob,
  };
}
