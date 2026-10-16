import React, { useState, useEffect, useCallback, useRef } from "react";
import {
  Box,
  Paper,
  Typography,
  Chip,
  Divider,
  Stack,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  Tabs,
  Tab,
  SelectChangeEvent,
} from "@mui/material";
import {
  Dns as PodIcon,
  Terminal as LogsIcon,
  FolderZip as FileIcon,
  CommitOutlined as CommitOutlinedIcon,
  CheckCircle as RunningIcon,
  Error as ErrorIcon,
  Pending as PendingIcon,
} from "@mui/icons-material";

import { useNamespaces } from "../hooks/useNamespaces";
import { usePods } from "../hooks/usePods";
import { ContainerFileExplorerTab } from "./ContainerFileExplorerTab";
import { RunJobListItem } from "./RunJobListItem";
import { PodListSidebar } from "./PodListSidebar";
import { PodStatus } from "../models/k8sResource";
import { ConsoleLogsTab } from "./ConsoleLogsTab.tsx";

const MIN_SIDEBAR_WIDTH = 240;
const MAX_SIDEBAR_WIDTH = 600;
const DEFAULT_SIDEBAR_WIDTH = 340;

export const PodOperationsView: React.FC = () => {
  const {
    namespaces,
    selectedNs,
    setSelectedNs,
    loading: loadingNs,
    error: nsError,
  } = useNamespaces();

  const [searchQuery, setSearchQuery] = useState<string>("");
  const [namespaceFilter, setNamespaceFilter] = useState<string>("all");

  const {
    pods,
    selectedPod,
    setSelectedPod,
    logs,
    loading: loadingPods,
    logsLoading,
    error: podsError,
    fetchPods,
    fetchLogs,
  } = usePods(namespaceFilter);

  const [podTab, setPodTab] = useState<number>(0);
  const [selectedContainerIndex, setSelectedContainerIndex] =
    useState<number>(0);

  // Resizable Sidebar State
  const [sidebarWidth, setSidebarWidth] = useState<number>(
    DEFAULT_SIDEBAR_WIDTH,
  );
  const [isResizing, setIsResizing] = useState<boolean>(false);
  const isResizingRef = useRef<boolean>(false);

  const startResizing = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsResizing(true);
    isResizingRef.current = true;
  }, []);

  const stopResizing = useCallback(() => {
    setIsResizing(false);
    isResizingRef.current = false;
  }, []);

  const resize = useCallback((e: MouseEvent) => {
    if (isResizingRef.current) {
      const newWidth = e.clientX - 16;
      if (newWidth >= MIN_SIDEBAR_WIDTH && newWidth <= MAX_SIDEBAR_WIDTH) {
        setSidebarWidth(newWidth);
      }
    }
  }, []);

  useEffect(() => {
    window.addEventListener("mousemove", resize);
    window.addEventListener("mouseup", stopResizing);
    return () => {
      window.removeEventListener("mousemove", resize);
      window.removeEventListener("mouseup", stopResizing);
    };
  }, [resize, stopResizing]);

  useEffect(() => {
    if (selectedNs) {
      setNamespaceFilter(selectedNs);
    }
  }, [selectedNs]);

  const activeContainer = selectedPod?.containers?.[selectedContainerIndex];

  useEffect(() => {
    if (selectedPod && activeContainer && podTab === 0) {
      fetchLogs(selectedPod.namespace, selectedPod.name, activeContainer.name);
    }
  }, [selectedPod, selectedContainerIndex, podTab, fetchLogs, activeContainer]);

  const handleNamespaceChange = (e: SelectChangeEvent) => {
    const val = e.target.value;
    setNamespaceFilter(val);
    if (val !== "all") {
      setSelectedNs(val);
    }
  };

  const filteredPods = (pods || []).filter((pod) =>
    pod.name.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const getStatusChip = (status: PodStatus) => {
    switch (status) {
      case "Running":
        return (
          <Chip
            icon={<RunningIcon />}
            label="Running"
            color="success"
            size="small"
            variant="outlined"
          />
        );
      case "Pending":
        return (
          <Chip
            icon={<PendingIcon />}
            label="Pending"
            color="warning"
            size="small"
            variant="outlined"
          />
        );
      case "CrashLoopBackOff":
      case "Failed":
        return (
          <Chip
            icon={<ErrorIcon />}
            label={status}
            color="error"
            size="small"
            variant="outlined"
          />
        );
      default:
        return <Chip label={status} size="small" variant="outlined" />;
    }
  };

  return (
    <Box
      sx={{
        display: "flex",
        flexGrow: 1,
        minHeight: 0,
        width: "100%",
        overflow: "hidden",
        userSelect: isResizing ? "none" : "auto",
      }}
    >
      {/* Resizable Left Sidebar */}
      <Box
        sx={{
          width: `${sidebarWidth}px`,
          minWidth: `${MIN_SIDEBAR_WIDTH}px`,
          maxWidth: `${MAX_SIDEBAR_WIDTH}px`,
          height: "100%",
          flexShrink: 0,
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
        }}
      >
        <PodListSidebar
          pods={filteredPods}
          selectedPod={selectedPod}
          namespaces={namespaces}
          namespaceFilter={namespaceFilter}
          searchQuery={searchQuery}
          loadingNs={loadingNs}
          loadingPods={loadingPods}
          displayError={nsError || podsError}
          onSearchChange={setSearchQuery}
          onNamespaceChange={handleNamespaceChange}
          onSelectPod={(pod) => {
            setSelectedPod(pod);
            setSelectedContainerIndex(0);
          }}
          onRefresh={fetchPods}
        />
      </Box>

      {/* Splitter Handle */}
      <Box
        onMouseDown={startResizing}
        sx={{
          width: "8px",
          cursor: "col-resize",
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          bgcolor: "transparent",
          transition: "background-color 0.2s",
          "&:hover": { bgcolor: "action.hover" },
          "&::after": {
            content: '""',
            width: "2px",
            height: "24px",
            borderRadius: "1px",
            bgcolor: isResizing ? "primary.main" : "divider",
          },
        }}
      />

      {/* Pod Details Container */}
      <Box
        sx={{
          flexGrow: 1,
          minWidth: 0,
          height: "100%",
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
        }}
      >
        <Paper
          elevation={0}
          variant="outlined"
          sx={{
            p: 2.5,
            height: "100%",
            display: "flex",
            flexDirection: "column",
            boxSizing: "border-box",
            borderColor: "divider",
            overflow: "hidden",
          }}
        >
          {selectedPod ? (
            <>
              <Box sx={{ flexShrink: 0 }}>
                <Stack
                  direction="row"
                  sx={{
                    justifyContent: "space-between",
                    alignItems: "flex-start",
                    mb: 2,
                  }}
                >
                  <Box>
                    <Stack
                      direction="row"
                      spacing={1}
                      sx={{ alignItems: "center", mb: 0.5 }}
                    >
                      <PodIcon color="primary" />
                      <Typography variant="h6" sx={{ fontWeight: 600 }}>
                        {selectedPod.name}
                      </Typography>
                    </Stack>
                    <Typography variant="caption" color="text.secondary">
                      Namespace: <strong>{selectedPod.namespace}</strong> | IP:{" "}
                      <strong>{selectedPod.ip}</strong>
                    </Typography>
                  </Box>
                  {getStatusChip(selectedPod.status)}
                </Stack>

                <Stack
                  direction="row"
                  spacing={2}
                  sx={{ alignItems: "center", mb: 1.5 }}
                >
                  <FormControl size="small" sx={{ minWidth: 180 }}>
                    <InputLabel id="container-select-label">
                      Container Target
                    </InputLabel>
                    <Select
                      labelId="container-select-label"
                      value={selectedContainerIndex.toString()}
                      label="Container Target"
                      onChange={(e: SelectChangeEvent) =>
                        setSelectedContainerIndex(Number(e.target.value))
                      }
                    >
                      {selectedPod.containers?.map((c, idx) => (
                        <MenuItem key={idx} value={idx}>
                          {c.name}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>

                  <Tabs value={podTab} onChange={(_, val) => setPodTab(val)}>
                    <Tab
                      icon={<LogsIcon fontSize="small" />}
                      iconPosition="start"
                      label="Console Logs"
                    />
                    <Tab
                      icon={<FileIcon fontSize="small" />}
                      iconPosition="start"
                      label="File Explorer"
                    />
                    <Tab
                      icon={<CommitOutlinedIcon fontSize="small" />}
                      iconPosition="start"
                      label="Quick Action"
                    />
                  </Tabs>
                </Stack>

                <Divider sx={{ mb: 2 }} />
              </Box>

              {/* Scrollable Pod Content */}
              <Box
                sx={{
                  flexGrow: 1,
                  minHeight: 0,
                  height: "100%",
                  overflow: "hidden",
                }}
              >
                {podTab === 0 && (
                  <ConsoleLogsTab logs={logs} logsLoading={logsLoading} />
                )}
                {podTab === 1 && (
                  <Box sx={{ height: "100%", overflow: "auto" }}>
                    <ContainerFileExplorerTab
                      podNamespace={selectedPod.namespace}
                      podName={selectedPod.name}
                      activeContainer={activeContainer}
                    />
                  </Box>
                )}
                {podTab === 2 && (
                  <Box sx={{ height: "100%", overflow: "auto" }}>
                    <RunJobListItem
                      podNamespace={selectedPod.namespace}
                      podName={selectedPod.name}
                      activeContainer={activeContainer}
                    />
                  </Box>
                )}
              </Box>
            </>
          ) : (
            <Stack
              sx={{
                alignItems: "center",
                justifyContent: "center",
                height: "100%",
                color: "text.secondary",
              }}
            >
              <PodIcon sx={{ fontSize: 56, mb: 1, opacity: 0.3 }} />
              <Typography variant="body2">
                Select an execution pod from the left panel
              </Typography>
            </Stack>
          )}
        </Paper>
      </Box>
    </Box>
  );
};
