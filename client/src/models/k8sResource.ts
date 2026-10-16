export interface NamespaceItem {
  name: string;
  status: string;
  age: string;
}

export interface NamespacesResponse {
  total: number;
  limit: number;
  data: NamespaceItem[];
}

export interface ContainerInfo {
  name: string;
  image: string;
  restarts: number;
  ready: boolean;
}

export interface ContainerInfo {
  name: string;
  image: string;
  restarts: number;
  ready: boolean;
  files?: ContainerFile[];
}

export interface ContainerItem {
  name: string;
  image: string;
  restarts: number;
  ready: boolean;
  cpuUsage?: string;
  memoryUsage?: string;
}
export interface PodItem {
  id: string;
  name: string;
  namespace: string;
  status: PodStatus;
  node: string;
  ip: string;
  age: string;
  cpuUsage?: string;
  memoryUsage?: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  conditions?: { type: string; status: string; lastTransitionTime?: string }[];
  containers: ContainerItem[];
}
export type PodStatus =
  | "Running"
  | "Pending"
  | "CrashLoopBackOff"
  | "Failed"
  | "Succeeded"
  | "Unknown";

export interface ContainerFile {
  name: string;
  path: string;
  isDirectory: boolean;
  size: number;
  type?: "code" | "config" | "log" | "script" | "text" | "unknown";
  permissions?: string;
  updatedAt?: string;
  content?: string;
  editable: boolean;
}

export interface ContainerFile {
  name: string;
  path: string;
  isDirectory: boolean;
  size: number;
  type?: "code" | "config" | "log" | "script" | "text" | "unknown";
  permissions?: string;
  updatedAt?: string;
  content?: string;
}

export type FileType = "doc" | "code" | "data" | "lib" | "unknow";

export interface FileItem {
  id: string;
  name: string;
  type: FileType;
  size: string;
  updatedAt: string;
  content: string;
}

export interface ContainerInfo {
  name: string;
  image: string;
  restarts: number;
  ready: boolean;
  files?: ContainerFile[];
}
