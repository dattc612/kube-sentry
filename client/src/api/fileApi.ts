import { ContainerFile } from "../models/k8sResource";
import { api } from "./apiClient";

interface FileContentResponse {
  data: string;
  error: string;
}

export const fileApi = {
  // GET /:namespace/:pod/containers/:container/files
  async listFiles(
    namespace: string,
    pod: string,
    container: string,
    path: string,
  ): Promise<ContainerFile[]> {
    const { data } = await api.get<ContainerFile[]>(
      `pods/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}/containers/${encodeURIComponent(container)}/files`,
      {
        params: { path },
      },
    );
    return data;
  },

  async getFileContent(
    namespace: string,
    podName: string,
    containerName: string,
    filePath: string,
  ): Promise<FileContentResponse> {
    const { data } = await api.get<FileContentResponse>(
      `pods/${encodeURIComponent(namespace)}/${encodeURIComponent(podName)}/containers/${encodeURIComponent(containerName)}/fileContent`,
      {
        params: { path: filePath },
        responseType: "json",
      },
    );
    return data;
  },

  downloadFile: async (
    namespace: string,
    pod: string,
    container: string,
    path: string,
    fileName: string,
    setDownloadedBytes: (bytes: number) => void,
  ): Promise<void> => {
    try {
      console.log(`Starting download request for: ${fileName}`);

      const response = await api.get(
        `/pods/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}/containers/${encodeURIComponent(container)}/download`,
        {
          params: { path },
          responseType: "blob",
          timeout: 0,
          onDownloadProgress: (progressEvent) => {
            setDownloadedBytes(progressEvent.loaded);
          },
        },
      );
      const contentType = String(
        response.headers["content-type"] || "application/octet-stream",
      );
      const blob = new Blob([response.data], { type: contentType });

      const downloadUrl = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = downloadUrl;
      link.download = fileName;

      document.body.appendChild(link);
      link.click();

      document.body.removeChild(link);
      window.URL.revokeObjectURL(downloadUrl);
    } catch (error) {
      console.error("The download request failed!: ", error);
    }
  },

  saveFileContent: async (
    namespace: string,
    pod: string,
    container: string,
    path: string,
    content: string,
  ): Promise<void> => {
    await api.put(
      `/pods/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}/containers/${encodeURIComponent(container)}/file`,
      { content },
      { params: { path } },
    );
  },
};
