import { TreeViewDefaultItemModelProperties } from "@mui/x-tree-view";

// TypeScript Interfaces
export interface FileItem {
  id: string;
  name: string;
  path: string;
  isDir: boolean;
  size: number;
}

export interface TreeFileItem extends TreeViewDefaultItemModelProperties {
  id: string;
  label: string;
  isDir: boolean;
  size: number;
  path: string;
  children?: TreeFileItem[];
}

export interface PodInfo {
  namespace: string;
  podName: string;
  container: string;
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
