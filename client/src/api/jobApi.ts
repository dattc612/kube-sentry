import { JobItem, JobItemResponse } from "../models/jobs";
import { api } from "./apiClient";

export interface RunJobResponse {
  output?: string;
  error?: string;
}

export const jobsApi = {
  // GET /:namespace/:pod/containers/:container/jobs
  async listJobs(
    namespace: string,
    pod: string,
    container: string,
  ): Promise<JobItem[]> {
    const { data } = (
      await api.get<JobItemResponse>(
        `/pods/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}/containers/${encodeURIComponent(container)}/jobs`,
      )
    ).data;
    return data;
  },

  // GET /:namespace/:pod/containers/:container/job/:id/run
  runJob: async (
    id: number,
    namespace: string,
    pod: string,
    container: string,
  ): Promise<RunJobResponse> => {
    const response = await api.post(
      `/pods/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}/containers/${encodeURIComponent(container)}/job/${encodeURIComponent(id)}/run`,
    );
    return response.data;
  },
};
