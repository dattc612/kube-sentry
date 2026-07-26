import { useState, useEffect } from "react";
import { namespaceApi } from "../api/namespaceApi";
import { NamespaceItem } from "../models/k8sResource";

export function useNamespaces() {
  const [namespaces, setNamespaces] = useState<NamespaceItem[]>([]);
  const [selectedNs, setSelectedNs] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadNamespaces() {
      try {
        setLoading(true);
        const response = await namespaceApi.getNamespaces();

        const list = response.data || [];
        setNamespaces(list);

      } catch (err: any) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }

    loadNamespaces();
  }, []);

  return { namespaces, selectedNs, setSelectedNs, loading, error };
}