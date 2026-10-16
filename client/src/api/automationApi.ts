import {
  Automation,
  AutomationResponse,
  ExecutionListResponse,
  ExecutionLog,
  ExecutionStatusResponse,
  ExecutionStepsResponse,
  TriggerJobResponse,
} from "../models/automation";
import { api } from "./apiClient";

export const automationApi = {
  // POST /api/v1/automations/:automation_code/run
  triggerJob: async (
    automationCode: string,
    payload?: Record<string, any> | string,
  ) => {
    const response = await api.post<TriggerJobResponse>(
      `/automations/${automationCode}/run`,
      payload ?? {},
    );
    return response.data.data;
  },

  // GET /api/v1/executions?limit=20&offset=0
  listExecutions: async (
    limit = 20,
    offset = 0,
    sort_by = "created_at",
    order: "asc" | "desc" = "desc",
  ) => {
    const response = await api.get<ExecutionListResponse>("/executions", {
      params: {
        limit,
        offset,
        sort_by,
        order,
      },
    });
    return response.data;
  },

  // GET /api/v1/executions/:id/steps
  getExecutionSteps: async (executionId: number) => {
    const response = await api.get<ExecutionStepsResponse>(
      `/executions/${executionId}/steps`,
    );
    return response.data.data;
  },

  // GET /api/v1/executions/:id/status (Lightweight polling endpoint)
  getExecutionStatus: async (executionId: number) => {
    const response = await api.get<ExecutionStatusResponse>(
      `/executions/${executionId}/status`,
    );
    return response.data;
  },

  // GET /api/v1/executions/:id/logs
  getExecutionLogs: async (executionId: number) => {
    const response = await api.get<{
      execution_id: number;
      logs: ExecutionLog[];
    }>(`/executions/${executionId}/logs`);
    return response.data.logs;
  },

  // GET /api/v1/automations
  getAutomations: async () => {
    const response = await api.get<AutomationResponse>("/automations");
    return response.data;
  },
};
