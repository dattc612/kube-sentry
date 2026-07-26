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

export interface PodItem {
	id: string;
	name: string;
	namespace: string;
	status: PodStatus;
	node: string;
	ip: string;
	age: string;
	cpuUsage: string;
	memoryUsage: string;
	containers: ContainerInfo[];
	logs: string;
}
export type PodStatus = 'Running' | 'Pending' | 'CrashLoopBackOff' | 'Completed' | 'Failed';

export interface ContainerFile {
	name: string;
	path: string;
	isDirectory: boolean;
	size?: string;
	type?: 'code' | 'config' | 'log' | 'script' | 'text' | 'unknown';
	permissions?: string;
	updatedAt?: string;
	content?: string;
	editable: boolean
}

export interface ContainerFile {
	name: string;
	path: string;
	isDirectory: boolean;
	size?: string;
	type?: 'code' | 'config' | 'log' | 'script' | 'text' | 'unknown';
	permissions?: string;
	updatedAt?: string;
	content?: string;
}

export type FileType = 'doc' | 'code' | 'data' | 'lib' | 'unknow';

export interface FileItem {
	id: string;
	name: string;
	type: FileType;
	size: string;
	updatedAt: string;
	content: string;
}
