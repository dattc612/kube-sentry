export interface Automation {
  id: number;
  code: string;
  name: string;
  description: string;
  type: "action" | "rule" | "workflow";
  pack: string;
  status: "active" | "disabled";
  sample_input?: {};
  created_at?: string;
  updated_at?: string;
}
export interface AutomationResponse {
  data: Automation[];
  limit: number;
  offset: number;
}

export type ExecutionStatus =
  | "queued"
  | "running"
  | "succeeded"
  | "failed"
  | "cancelled";

export interface Execution {
  id: number;
  name: string;
  execution_code: string;
  automation_code: number;
  trigger_id?: number;
  trigger_source: string;
  status: ExecutionStatus;
  started_at?: string;
  finished_at?: string;
  duration_ms?: number;
  triggered_by?: string;
  input?: string;
  output?: string;
  error?: string;
  created_at: string;
}

export interface ExecutionStep {
  id: number;
  execution_id: number;
  step_number: number;
  name: string;
  status: ExecutionStatus;
  started_at?: string;
  finished_at?: string;
  duration_ms?: number;
  input?: string;
  output?: string;
  error?: string;
  log?: ExecutionLog;
}

export interface ExecutionLog {
  id: number;
  execution_id: number;
  step_id?: number;
  level: "debug" | "info" | "warn" | "error";
  message: string;
  created_at: string;
}

// Response Wrappers
export interface TriggerJobResponse {
  data: {
    execution_id: number;
    execution_code: string;
    status: ExecutionStatus;
    created_at: string;
  };
}

export interface ExecutionStepsResponse {
  data: ExecutionStep[];
}

export interface ExecutionStatusResponse {
  id: number;
  execution_code: string;
  status: ExecutionStatus;
  duration_ms?: number;
  started_at?: string;
  finished_at?: string;
  error?: string;
}

export interface ExecutionListResponse {
  data: Execution[];
  limit: number;
  offset: number;
  total: number;
}
