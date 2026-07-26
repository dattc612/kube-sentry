import { useState, useEffect, useCallback } from 'react';
import { ContainerFile } from '../models/k8sResource';
import { fileApi } from '../api/fileApi';

interface UseFilesParams {
  namespace: string;
  pod: string;
  container: string;
  path: string;
}

export function useFiles({ namespace, pod, container, path }: UseFilesParams) {
  const [files, setFiles] = useState<ContainerFile[]>([]);
  const [fileContent, setFileContent] = useState<string>('');
  const [loadingFiles, setLoadingFiles] = useState<boolean>(false);
  const [loadingContent, setLoadingContent] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const fetchFiles = useCallback(async () => {
    if (!namespace || !pod || !container) return;
    try {
      setLoadingFiles(true);
      setError(null);
      const data = await fileApi.listFiles(namespace, pod, container, path);
      setFiles(data);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch directory contents');
    } finally {
      setLoadingFiles(false);
    }
  }, [namespace, pod, container, path]);

  const fetchFileContent = useCallback(
    async (filePath: string) => {
      if (!namespace || !pod || !container || !filePath) return;
      try {
        setLoadingContent(true);
        const content = await fileApi.getFileContent(namespace, pod, container, filePath);
        setFileContent(content);
      } catch (err: any) {
        setFileContent(`Failed to load file content: ${err.message}`);
      } finally {
        setLoadingContent(false);
      }
    },
    [namespace, pod, container]
  );

	const [saving, setSaving] = useState(false);

	const saveFile = async (filePath: string, content: string): Promise<boolean> => {
		setSaving(true);
		try {
			await fileApi.saveFileContent(namespace, pod, container, filePath, content);
			setSaving(false);
			return true;
		} catch (err: any) {
			setError(err.message || 'Failed to save file');
			setSaving(false);
			return false;
		}
	};


  useEffect(() => {
    fetchFiles();
  }, [fetchFiles]);

	return {
		files,
		fileContent,
		loadingFiles,
		loadingContent,
		saving,
		error,
		fetchFiles,
		fetchFileContent,
		saveFile
	};
}

