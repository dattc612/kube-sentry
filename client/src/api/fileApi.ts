import { ContainerFile } from "../models/k8sResource";
import { api } from "./apiClient";

export const fileApi = {
	// GET /:namespace/:pod/containers/:container/files
	async listFiles(
  	namespace: string,
  	pod: string,
  	container: string,
  	path: string 
	): Promise<ContainerFile[]> {
  	const { data } = await api.get<ContainerFile[]>(
			`pods/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}/containers/${encodeURIComponent(container)}/files`,
			{
				params: { path },
			});
  	return data;
	},

  async getFileContent(
    namespace: string,
    podName: string,
    containerName: string,
    filePath: string
  ): Promise<string> {
    const { data } = await api.get<string>(`pods/${encodeURIComponent(namespace)}/${encodeURIComponent(podName)}/containers/${encodeURIComponent(containerName)}/fileContent`, {
      params: { path: filePath },
      responseType: "text",
    });
    return data;
  },

  downloadFile: async (
    namespace: string,
    pod: string,
    container: string,
    path: string,
    fileName: string
  ): Promise<void> => {
    const response = await api.get(
      `/pods/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}/containers/${encodeURIComponent(container)}/download`,
      {
        params: { path },
        responseType: 'blob'
      }
    );

    const contentType = String(response.headers['content-type'] || 'application/octet-stream');
		const blob = new Blob([response.data], {
			type: contentType
		});

    const downloadUrl = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = fileName;

    document.body.appendChild(link);
    link.click();

    document.body.removeChild(link);
    window.URL.revokeObjectURL(downloadUrl);
  },

  saveFileContent: async (
    namespace: string,
    pod: string,
    container: string,
    path: string,
    content: string
  ): Promise<void> => {
    await api.put(
      `/pods/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}/containers/${encodeURIComponent(container)}/file`,
      { content },
      { params: { path } }
    );
  }
};