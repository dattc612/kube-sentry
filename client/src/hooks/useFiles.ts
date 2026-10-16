import { useState, useEffect, useCallback, useRef } from "react";
import { ContainerFile } from "../models/k8sResource";
import { fileApi } from "../api/fileApi";

interface UseFilesParams {
  namespace: string;
  pod: string;
  container: string;
}

export function useFiles({
  namespace,
  pod: podName,
  container: containerName,
}: UseFilesParams) {
  const [files, setFiles] = useState<ContainerFile[]>([]);
  const [fileContent, setFileContent] = useState<string>("");
  const [loadingFiles, setLoadingFiles] = useState<boolean>(false);
  const [loadingContent, setLoadingContent] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPath, setCurrentPath] = useState<string>("/home");
  const [selectedFile, setSelectedFile] = useState<ContainerFile | null>(null);

  // Download state
  const [isDownloading, setIsDownloading] = useState<boolean>(false);
  const [downloadProgress, setDownloadProgress] = useState<number>(0);
  const [downloadedBytes, setDownloadedBytes] = useState<number>(0);
  const [saving, setSaving] = useState<boolean>(false);

  const fetchFiles = useCallback(async () => {
    if (!namespace || !podName || !containerName) return;
    try {
      setLoadingFiles(true);
      setError(null);
      const data = await fileApi.listFiles(
        namespace,
        podName,
        containerName,
        currentPath,
      );
      setFiles(data);
    } catch (err: any) {
      setError(err.message || "Failed to fetch directory contents");
    } finally {
      setLoadingFiles(false);
    }
  }, [currentPath, namespace, podName, containerName]);

  const handleDownload = useCallback(
    async (file: ContainerFile) => {
      if (file.isDirectory || !containerName) return;

      setSelectedFile(file);
      setIsDownloading(true);
      setDownloadProgress(0);
      setDownloadedBytes(0);

      try {
        await fileApi.downloadFile(
          namespace,
          podName,
          containerName,
          file.path,
          file.name,
          setDownloadedBytes,
        );
        setDownloadProgress(100);
      } catch (err: any) {
        setError(err.message || "Failed to download file");
      } finally {
        setIsDownloading(false);
      }
    },
    [namespace, podName, containerName],
  );

  useEffect(() => {
    if (isDownloading && selectedFile?.size && selectedFile.size > 0) {
      const progress = Math.min(
        (downloadedBytes / selectedFile.size) * 100,
        100,
      );
      setDownloadProgress(progress);
    }
  }, [downloadedBytes, isDownloading, selectedFile]);

  const fetchFileContent = useCallback(
    async (filePath: string) => {
      if (!namespace || !podName || !containerName || !filePath) return;
      try {
        setLoadingContent(true);
        const content = await fileApi.getFileContent(
          namespace,
          podName,
          containerName,
          filePath,
        );
        if (content.error) {
          setFileContent(`Failed to load file content: ${content.error}`);
        } else {
          setFileContent(content.data);
        }
      } catch (err: any) {
        setFileContent(`Failed to load file content: ${err.message}`);
      } finally {
        setLoadingContent(false);
      }
    },
    [selectedFile],
  );

  const saveFile = useCallback(
    async (filePath: string, content: string): Promise<boolean> => {
      setSaving(true);
      try {
        await fileApi.saveFileContent(
          namespace,
          podName,
          containerName,
          filePath,
          content,
        );
        return true;
      } catch (err: any) {
        setError(err.message || "Failed to save file");
        return false;
      } finally {
        setSaving(false);
      }
    },
    [namespace, podName, containerName, selectedFile],
  );

  //  fix for the issue where the file list is refreshed twice when the namespace, pod, or container changes
  const isResettingRef = useRef(false);

  useEffect(() => {
    isResettingRef.current = true;
    setCurrentPath("/");
  }, [namespace, podName, containerName]);

  useEffect(() => {
    if (isResettingRef.current && currentPath !== "/home") {
      return;
    }
    isResettingRef.current = false;
    fetchFiles();
  }, [fetchFiles, currentPath]);

  return {
    files,
    fileContent,
    loadingFiles,
    loadingContent,
    saving,
    error,
    fetchFiles,
    fetchFileContent,
    saveFile,
    handleDownload,
    isDownloading,
    downloadProgress,
    currentPath,
    setCurrentPath,
    selectedFile,
    setSelectedFile,
  };
}
