import React, { useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  FormControl,
  IconButton,
  InputAdornment,
  InputLabel,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  ListSubheader,
  MenuItem,
  Paper,
  Select,
  Snackbar,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import {
  Dns as PodIcon,
  Layers as DeploymentIcon,
  Timer as CronJobIcon,
  NetworkCheck as ServiceIcon,
  AltRoute as IngressIcon,
  Storage as StorageIcon,
  FolderZip as VolumeIcon,
  SettingsApplications as ConfigMapIcon,
  VpnKey as SecretIcon,
  Security as RbacIcon,
  Computer as NodeIcon,
  CloudQueue as NamespaceIcon,
  Search as SearchIcon,
  Refresh as RefreshIcon,
  Edit as EditIcon,
  Visibility as VisibilityIcon,
  Save as SaveIcon,
  Close as CloseIcon,
} from "@mui/icons-material";
import Editor from "@monaco-editor/react";
import { parse, stringify } from "yaml";
import { useK8sResources } from "../hooks/useK8sResources";
import { k8sApi } from "../api/k8sApi";
import { useNamespaces } from "../hooks/useNamespaces";

export interface ColumnConfig {
  id: string;
  label: string;
  accessor?: (row: any) => any;
  render?: (val: any, row: any) => React.ReactNode;
}

interface ResourceItem {
  id: string;
  name: string;
  label: string;
  icon: React.ReactNode;
  category: string;
  columns: ColumnConfig[];
}

// Shared visual helpers for consistent, colorful status and namespace indicators.
const chipSx = {
  fontWeight: 600,
  fontSize: "0.75rem",
  height: 23,
  borderRadius: "6px",
};

const namespaceChip = (value: unknown) =>
  value ? (
    <Chip
      label={String(value)}
      size="small"
      color="primary"
      variant="outlined"
      sx={chipSx}
    />
  ) : (
    "-"
  );

const ReadyChip = ({ value }: { value: unknown }) => {
  const [ready, total] = String(value ?? "0/0")
    .split("/")
    .map(Number);
  const complete = total > 0 && ready === total;
  return (
    <Chip
      label={`${ready}/${total}`}
      size="small"
      color={complete ? "success" : "warning"}
      sx={chipSx}
    />
  );
};

const StatusChip = ({ value }: { value: unknown }) => {
  const status = String(value ?? "Unknown");
  const color = [
    "Running",
    "Ready",
    "Bound",
    "Active",
    "Succeeded",
    "Available",
  ].includes(status)
    ? "success"
    : ["Pending", "Terminating", "ContainerCreating", "Released"].includes(
          status,
        )
      ? "warning"
      : [
            "Failed",
            "Error",
            "CrashLoopBackOff",
            "ImagePullBackOff",
            "ErrImagePull",
            "Lost",
            "NotReady",
          ].includes(status)
        ? "error"
        : status === "Completed"
          ? "info"
          : "default";
  return (
    <Chip
      label={status}
      size="small"
      color={color}
      variant={color === "default" ? "outlined" : "filled"}
      sx={chipSx}
    />
  );
};

// ISO timestamp -> compact relative age. Existing API-provided age values take precedence.
const formatResourceAge = (timestamp?: string) => {
  if (!timestamp) return "-";
  const elapsed = Date.now() - new Date(timestamp).getTime();
  if (!Number.isFinite(elapsed) || elapsed < 0) return "-";
  const minutes = Math.floor(elapsed / 60000);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.floor(hours / 24);
  if (days < 365) return `${days}d`;
  return `${Math.floor(days / 365)}y`;
};

const RESOURCE_TYPES: ResourceItem[] = [
  // Workloads
  {
    id: "pods",
    name: "Pods",
    label: "Pods",
    icon: <PodIcon fontSize="small" />,
    category: "Workloads",
    columns: [
      { id: "name", label: "Name" },
      { id: "namespace", label: "Namespace", render: (v) => namespaceChip(v) },
      {
        id: "ready",
        label: "Ready",
        accessor: (row) => {
          const containers = Array.isArray(row.containers)
            ? row.containers
            : [];
          const ready = containers.filter(
            (container: any) => container.ready,
          ).length;
          return `${ready}/${containers.length}`;
        },
        render: (v) => {
          const [ready, total] = String(v || "0/0")
            .split("/")
            .map(Number);
          return (
            <Chip
              label={`${ready}/${total}`}
              size="small"
              color={total > 0 && ready === total ? "success" : "warning"}
              sx={chipSx}
            />
          );
        },
      },
      {
        id: "status",
        label: "Status",
        render: (v) => <StatusChip value={v || "Unknown"} />,
      },
      {
        id: "restarts",
        label: "Restarts",
        accessor: (row) =>
          Array.isArray(row.containers)
            ? row.containers.reduce(
                (sum: number, c: any) => sum + (c.restarts || 0),
                0,
              )
            : 0,
        render: (v) => (
          <Chip
            label={v ?? 0}
            size="small"
            color={Number(v) > 0 ? "error" : "success"}
            variant="outlined"
            sx={chipSx}
          />
        ),
      },
      { id: "age", label: "Age" },
      { id: "node", label: "Node" },
      { id: "ip", label: "IP Address" },
      { id: "cpuUsage", label: "CPU Usage" },
      { id: "memoryUsage", label: "Memory Usage" },
    ],
  },
  {
    id: "deployments",
    name: "Deployments",
    label: "Deployments",
    icon: <DeploymentIcon fontSize="small" />,
    category: "Workloads",
    columns: [
      { id: "name", label: "Name" },
      { id: "namespace", label: "Namespace", render: (v) => namespaceChip(v) },
      { id: "ready", label: "Ready", render: (v) => <ReadyChip value={v} /> },
      { id: "upToDate", label: "Up-to-date" },
      { id: "available", label: "Available" },
      { id: "age", label: "Age" },
    ],
  },
  {
    id: "statefulsets",
    name: "StatefulSets",
    label: "StatefulSets",
    icon: <DeploymentIcon fontSize="small" />,
    category: "Workloads",
    columns: [
      { id: "name", label: "Name" },
      { id: "namespace", label: "Namespace", render: (v) => namespaceChip(v) },
      { id: "ready", label: "Ready", render: (v) => <ReadyChip value={v} /> },
      { id: "replicas", label: "Replicas" },
      { id: "age", label: "Age" },
    ],
  },
  {
    id: "daemonsets",
    name: "DaemonSets",
    label: "DaemonSets",
    icon: <DeploymentIcon fontSize="small" />,
    category: "Workloads",
    columns: [
      { id: "name", label: "Name" },
      { id: "namespace", label: "Namespace", render: (v) => namespaceChip(v) },
      { id: "desired", label: "Desired" },
      { id: "current", label: "Current" },
      { id: "ready", label: "Ready", render: (v) => <ReadyChip value={v} /> },
      { id: "available", label: "Available" },
      { id: "age", label: "Age" },
    ],
  },
  {
    id: "cronjobs",
    name: "CronJobs",
    label: "CronJobs",
    icon: <CronJobIcon fontSize="small" />,
    category: "Workloads",
    columns: [
      {
        id: "name",
        label: "Name",
        accessor: (row) => row.metadata?.name ?? row.name,
      },
      {
        id: "namespace",
        label: "Namespace",
        accessor: (row) => row.metadata?.namespace ?? row.namespace,
        render: (v) => namespaceChip(v),
      },
      {
        id: "schedule",
        label: "Schedule",
        accessor: (row) => row.spec?.schedule ?? "-",
      },
      {
        id: "suspend",
        label: "Suspended",
        accessor: (row) => (row.spec?.suspend ? "Yes" : "No"),
        render: (v) => (
          <Chip
            label={v}
            size="small"
            color={v === "Yes" ? "warning" : "success"}
            sx={chipSx}
          />
        ),
      },
      {
        id: "active",
        label: "Active",
        accessor: (row) => row.status?.active?.length ?? 0,
      },
      {
        id: "age",
        label: "Age",
        accessor: (row) =>
          row.age ?? formatResourceAge(row.metadata?.creationTimestamp),
      },
    ],
  },

  // Networking
  {
    id: "services",
    name: "Services",
    label: "Services",
    icon: <ServiceIcon fontSize="small" />,
    category: "Network",
    columns: [
      { id: "name", label: "Name" },
      { id: "namespace", label: "Namespace", render: (v) => namespaceChip(v) },
      {
        id: "type",
        label: "Type",
        render: (v) => (
          <Chip
            label={v || "Unknown"}
            size="small"
            color={
              v === "LoadBalancer"
                ? "success"
                : v === "NodePort"
                  ? "info"
                  : "default"
            }
            variant="outlined"
            sx={chipSx}
          />
        ),
      },
      { id: "clusterIp", label: "Cluster IP" },
      { id: "externalIp", label: "External IP" },
      { id: "ports", label: "Port(s)" },
      { id: "age", label: "Age" },
    ],
  },
  {
    id: "ingresses",
    name: "Ingresses",
    label: "Ingresses",
    icon: <IngressIcon fontSize="small" />,
    category: "Network",
    columns: [
      { id: "name", label: "Name" },
      { id: "namespace", label: "Namespace", render: (v) => namespaceChip(v) },
      { id: "class", label: "Class", render: (v) => v || "-" },
      { id: "hosts", label: "Hosts", render: (v) => v || "*" },
      { id: "address", label: "Address", render: (v) => v || "-" },
      { id: "ports", label: "Ports", render: (v) => v || "-" },
      { id: "age", label: "Age" },
    ],
  },

  // Storage
  {
    id: "pvcs",
    name: "PersistentVolumeClaims",
    label: "PVCs",
    icon: <VolumeIcon fontSize="small" />,
    category: "Storage",
    columns: [
      {
        id: "name",
        label: "Name",
        accessor: (row) => row.metadata?.name ?? row.name,
      },
      {
        id: "namespace",
        label: "Namespace",
        accessor: (row) => row.metadata?.namespace ?? row.namespace,
        render: (v) => namespaceChip(v),
      },
      {
        id: "status",
        label: "Status",
        accessor: (row) => row.status?.phase ?? "Unknown",
        render: (v) => <StatusChip value={v} />,
      },
      {
        id: "volume",
        label: "Volume",
        accessor: (row) => row.spec?.volumeName ?? "-",
      },
      {
        id: "capacity",
        label: "Capacity",
        accessor: (row) =>
          row.status?.capacity?.storage ??
          row.spec?.resources?.requests?.storage ??
          "-",
      },
      {
        id: "storageClass",
        label: "StorageClass",
        accessor: (row) => row.spec?.storageClassName ?? "-",
      },
      {
        id: "age",
        label: "Age",
        accessor: (row) =>
          row.age ?? formatResourceAge(row.metadata?.creationTimestamp),
      },
    ],
  },
  {
    id: "pvs",
    name: "PersistentVolumes",
    label: "PVs",
    icon: <StorageIcon fontSize="small" />,
    category: "Storage",
    columns: [
      {
        id: "name",
        label: "Name",
        accessor: (row) => row.metadata?.name ?? row.name,
      },
      {
        id: "capacity",
        label: "Capacity",
        accessor: (row) => row.spec?.capacity?.storage ?? "-",
      },
      {
        id: "accessModes",
        label: "Access Modes",
        accessor: (row) => row.spec?.accessModes?.join(", ") ?? "-",
      },
      {
        id: "reclaimPolicy",
        label: "Reclaim Policy",
        accessor: (row) => row.spec?.persistentVolumeReclaimPolicy ?? "-",
        render: (v) => (
          <Chip
            label={v}
            size="small"
            color={
              v === "Retain" ? "warning" : v === "Delete" ? "error" : "default"
            }
            variant="outlined"
            sx={chipSx}
          />
        ),
      },
      {
        id: "status",
        label: "Status",
        accessor: (row) => row.status?.phase ?? "Unknown",
        render: (v) => <StatusChip value={v} />,
      },
      {
        id: "storageClass",
        label: "StorageClass",
        accessor: (row) => row.spec?.storageClassName ?? "-",
      },
    ],
  },

  // Configuration
  {
    id: "configmaps",
    name: "ConfigMaps",
    label: "ConfigMaps",
    icon: <ConfigMapIcon fontSize="small" />,
    category: "Config",
    columns: [
      { id: "name", label: "Name" },
      { id: "namespace", label: "Namespace", render: (v) => namespaceChip(v) },
      { id: "keysCount", label: "Data Keys" },
      { id: "age", label: "Age" },
    ],
  },
  {
    id: "secrets",
    name: "Secrets",
    label: "Secrets",
    icon: <SecretIcon fontSize="small" />,
    category: "Config",
    columns: [
      { id: "name", label: "Name" },
      { id: "namespace", label: "Namespace", render: (v) => namespaceChip(v) },
      {
        id: "type",
        label: "Type",
        render: (v) => (
          <Chip
            label={v || "Opaque"}
            size="small"
            color="warning"
            variant="outlined"
            sx={chipSx}
          />
        ),
      },
      { id: "keysCount", label: "Data Keys" },
      { id: "age", label: "Age" },
    ],
  },

  // Access control
  {
    id: "serviceaccounts",
    name: "ServiceAccounts",
    label: "ServiceAccounts",
    icon: <RbacIcon fontSize="small" />,
    category: "Access Control",
    columns: [
      { id: "name", label: "Name" },
      { id: "namespace", label: "Namespace", render: (v) => namespaceChip(v) },
      { id: "secrets", label: "Secrets" },
      { id: "age", label: "Age" },
    ],
  },
  {
    id: "roles",
    name: "Roles & ClusterRoles",
    label: "Roles",
    icon: <RbacIcon fontSize="small" />,
    category: "Access Control",
    columns: [
      { id: "name", label: "Name" },
      { id: "namespace", label: "Namespace", render: (v) => namespaceChip(v) },
      {
        id: "kind",
        label: "Kind",
        render: (v) => (
          <Chip
            label={v || "Role"}
            size="small"
            color={v === "ClusterRole" ? "secondary" : "info"}
            variant="outlined"
            sx={chipSx}
          />
        ),
      },
      { id: "age", label: "Age" },
    ],
  },

  // Cluster
  {
    id: "nodes",
    name: "Nodes",
    label: "Nodes",
    icon: <NodeIcon fontSize="small" />,
    category: "Cluster",
    columns: [
      { id: "name", label: "Name" },
      {
        id: "status",
        label: "Status",
        render: (v) => <StatusChip value={v} />,
      },
      {
        id: "roles",
        label: "Roles",
        render: (v) => (
          <Chip
            label={v || "worker"}
            size="small"
            color="info"
            variant="outlined"
            sx={chipSx}
          />
        ),
      },
      { id: "version", label: "Version" },
      { id: "internalIp", label: "Internal IP" },
      { id: "age", label: "Age" },
    ],
  },
  {
    id: "namespaces",
    name: "Namespaces",
    label: "Namespaces",
    icon: <NamespaceIcon fontSize="small" />,
    category: "Cluster",
    columns: [
      { id: "name", label: "Name" },
      {
        id: "status",
        label: "Status",
        render: (v) => <StatusChip value={v} />,
      },
      { id: "age", label: "Age" },
    ],
  },
];

export type EditorMode = "view" | "edit";

export interface ResourceEditorState {
  mode: EditorMode;
  resourceType: string;
  resourceName: string;
  namespace?: string;
}

export interface K8sResourceManagementProps {
  onSaveResource?: (
    resourceType: string,
    resource: Record<string, unknown>,
    yamlText: string,
  ) => Promise<void>;
  onLoadResource?: (
    resourceType: string,
    name: string,
    namespace?: string,
  ) => Promise<Record<string, unknown>>;
}

type ResourceRow = Record<string, any>;

export const K8sResourceManagement: React.FC<K8sResourceManagementProps> = ({
  onSaveResource,
  onLoadResource,
}) => {
  const [selectedResourceType, setSelectedResourceType] =
    useState<ResourceItem>(RESOURCE_TYPES[0]);
  const [searchQuery, setSearchQuery] = useState("");
  const [namespaceFilter, setNamespaceFilter] = useState("all");

  const [editorOpen, setEditorOpen] = useState(false);
  const [editorMode, setEditorMode] = useState<EditorMode>("view");
  const [editorTitle, setEditorTitle] = useState("");
  const [editorText, setEditorText] = useState("");
  const [originalText, setOriginalText] = useState("");
  const [editorError, setEditorError] = useState("");
  const [saving, setSaving] = useState(false);
  const [loadingEditor, setLoadingEditor] = useState(false);
  const [notice, setNotice] = useState("");
  const {
    namespaces,
    selectedNs,
    setSelectedNs,
    loading: loadingNs,
    error: nsError,
  } = useNamespaces();

  const {
    data: rawData,
    loading,
    error,
    refetch,
  } = useK8sResources(selectedResourceType.id, namespaceFilter);

  const categories = Array.from(new Set(RESOURCE_TYPES.map((r) => r.category)));

  const filteredData = (rawData || []).filter((item: ResourceRow) =>
    String(item.name ?? item.metadata?.name ?? "")
      .toLowerCase()
      .includes(searchQuery.toLowerCase()),
  );

  const openResourceEditor = async (row: ResourceRow, mode: EditorMode) => {
    const resourceName = String(row.metadata?.name ?? row.name ?? "resource");
    const namespaceValue = row.metadata?.namespace ?? row.namespace;
    const namespace = namespaceValue ? String(namespaceValue) : undefined;

    setEditorMode(mode);
    setEditorTitle(namespace ? `${resourceName} (${namespace})` : resourceName);
    setEditorError("");
    setEditorOpen(true);
    setLoadingEditor(true);

    try {
      let yamlText: string;
      if (onLoadResource) {
        const resource = await onLoadResource(
          selectedResourceType.id,
          resourceName,
          namespace,
        );
        yamlText = stringify(resource);
      } else {
        // Fetch raw JSON object from API and convert to YAML for editing/viewing
        const resourceObj = await k8sApi.getResourceJSON(
          selectedResourceType.id,
          namespace || "default",
          resourceName,
        );
        yamlText = stringify(resourceObj);
      }
      setEditorText(yamlText);
      setOriginalText(yamlText);
    } catch (err) {
      setEditorError(
        err instanceof Error ? err.message : "Failed to load resource.",
      );
      setEditorText("");
      setOriginalText("");
    } finally {
      setLoadingEditor(false);
    }
  };

  const closeResourceEditor = () => {
    if (saving) return;
    setEditorOpen(false);
    setEditorError("");
  };

  const saveResource = async () => {
    try {
      setEditorError("");
      let parsed: unknown;
      try {
        parsed = parse(editorText);
      } catch (e) {
        throw new Error("Invalid YAML syntax.");
      }

      if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
        throw new Error(
          "The YAML document must contain a valid Kubernetes resource object.",
        );
      }

      const resource = parsed as Record<string, unknown>;
      const metadata = resource.metadata as
        | { name?: string; namespace?: string }
        | undefined;
      if (!metadata?.name) {
        throw new Error("The resource must include metadata.name.");
      }

      setSaving(true);
      if (onSaveResource) {
        await onSaveResource(selectedResourceType.id, resource, editorText);
      } else {
        const namespace = String(metadata?.namespace ?? namespaceFilter);
        if (!namespace || namespace === "all") {
          throw new Error(
            "Cannot save this resource because metadata.namespace is missing or ambiguous.",
          );
        }
        // Send as JSON object to backend endpoint
        await k8sApi.updateResourceJSON(
          selectedResourceType.id,
          namespace,
          metadata.name,
          resource,
        );
      }
      setEditorOpen(false);
      setNotice("Resource saved successfully.");
      await refetch();
    } catch (err) {
      setEditorError(
        err instanceof Error ? err.message : "Failed to save resource.",
      );
    } finally {
      setSaving(false);
    }
  };

  return (
    <Box
      sx={{
        flexGrow: 1,
        minHeight: 0,
        height: "100%",
        display: "flex",
        overflow: "hidden",
      }}
    >
      <Paper
        elevation={0}
        variant="outlined"
        sx={{
          width: 240,
          flexShrink: 0,
          mr: 1.5,
          height: "100%",
          display: "flex",
          flexDirection: "column",
          overflowY: "auto",
        }}
      >
        <List dense disablePadding sx={{ py: 1 }}>
          {categories.map((cat) => (
            <React.Fragment key={cat}>
              <ListSubheader
                sx={{
                  lineHeight: "28px",
                  fontSize: "0.75rem",
                  fontWeight: 700,
                  textTransform: "uppercase",
                  color: "text.secondary",
                  bgcolor: "background.paper",
                }}
              >
                {cat}
              </ListSubheader>
              {RESOURCE_TYPES.filter((r) => r.category === cat).map((res) => (
                <ListItemButton
                  key={res.id}
                  selected={selectedResourceType.id === res.id}
                  onClick={() => {
                    setSelectedResourceType(res);
                    setSearchQuery("");
                  }}
                  sx={{ py: 0.5, px: 2, borderRadius: 1, mx: 1, my: 0.2 }}
                >
                  <ListItemIcon
                    sx={{
                      minWidth: 28,
                      color:
                        selectedResourceType.id === res.id
                          ? "primary.main"
                          : "inherit",
                    }}
                  >
                    {res.icon}
                  </ListItemIcon>
                  <ListItemText
                    primary={res.name}
                    slotProps={{
                      primary: {
                        sx: {
                          fontSize: "0.85rem",
                          fontWeight:
                            selectedResourceType.id === res.id ? 600 : 400,
                        },
                      },
                    }}
                  />
                </ListItemButton>
              ))}
              <Divider sx={{ my: 1 }} />
            </React.Fragment>
          ))}
        </List>
      </Paper>

      <Paper
        elevation={0}
        variant="outlined"
        sx={{
          flexGrow: 1,
          minWidth: 0,
          height: "100%",
          display: "flex",
          flexDirection: "column",
          p: 2,
          boxSizing: "border-box",
          overflow: "hidden",
        }}
      >
        <Stack
          direction="row"
          justifyContent={"space-between" as const}
          alignItems={"center" as const}
          sx={{ mb: 2 }}
        >
          <Stack direction="row" spacing={1} alignItems={"center" as const}>
            {selectedResourceType.icon}
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              {selectedResourceType.name} ({filteredData.length})
            </Typography>
          </Stack>
          <Tooltip title="Refresh Cluster Resources">
            <span>
              <IconButton
                size="small"
                onClick={() => void refetch()}
                disabled={loading}
              >
                <RefreshIcon fontSize="small" />
              </IconButton>
            </span>
          </Tooltip>
        </Stack>

        <Stack
          direction="row"
          spacing={2}
          sx={{ mb: 2, flexShrink: 0, alignItems: "center" }}
        >
          <FormControl size="small" sx={{ minWidth: 160 }}>
            <InputLabel id="mgmt-namespace-label">Namespace</InputLabel>
            <Select
              labelId="mgmt-namespace-label"
              value={namespaceFilter}
              label="Namespace"
              onChange={(e) => setNamespaceFilter(e.target.value)}
            >
              <MenuItem value="all">All</MenuItem>
              {namespaces.map((ns) => (
                <MenuItem key={ns.name} value={ns.name}>
                  {ns.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          <TextField
            size="small"
            placeholder={`Search ${selectedResourceType.name.toLowerCase()}...`}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            sx={{ flexGrow: 1, maxWidth: 360 }}
            slotProps={{
              input: {
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon fontSize="small" color="action" />
                  </InputAdornment>
                ),
              },
            }}
          />
        </Stack>

        <Box
          sx={{
            flexGrow: 1,
            minHeight: 0,
            overflow: "auto",
            border: 1,
            borderColor: "divider",
            borderRadius: 1,
            position: "relative",
          }}
        >
          {loading ? (
            <Stack
              sx={{
                alignItems: "center",
                justifyContent: "center",
                height: "100%",
                py: 8,
              }}
            >
              <CircularProgress size={32} />
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Loading {selectedResourceType.name.toLowerCase()}...
              </Typography>
            </Stack>
          ) : error ? (
            <Stack
              sx={{
                alignItems: "center",
                justifyContent: "center",
                height: "100%",
                p: 3,
              }}
            >
              <Typography variant="body2" color="error">
                {error}
              </Typography>
            </Stack>
          ) : (
            <TableContainer sx={{ maxHeight: "100%" }}>
              <Table size="small" stickyHeader>
                <TableHead>
                  <TableRow>
                    {selectedResourceType.columns.map((col) => (
                      <TableCell key={col.id}>{col.label}</TableCell>
                    ))}
                    <TableCell align="right">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {filteredData.map((row: ResourceRow, rIdx: number) => (
                    <TableRow
                      key={
                        row.id ??
                        row.metadata?.uid ??
                        `${row.namespace ?? "cluster"}-${row.name ?? rIdx}`
                      }
                      hover
                    >
                      {selectedResourceType.columns.map((col) => {
                        const value = col.accessor
                          ? col.accessor(row)
                          : row[col.id];
                        return (
                          <TableCell
                            key={col.id}
                            sx={
                              col.id === "name"
                                ? { fontWeight: 500 }
                                : undefined
                            }
                          >
                            {col.render
                              ? col.render(value, row)
                              : typeof value === "object" && value !== null
                                ? JSON.stringify(value)
                                : (value ?? "-")}
                          </TableCell>
                        );
                      })}
                      <TableCell align="right" sx={{ whiteSpace: "nowrap" }}>
                        <Tooltip title="View YAML">
                          <IconButton
                            size="small"
                            onClick={() => void openResourceEditor(row, "view")}
                            aria-label="View YAML"
                          >
                            <VisibilityIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Edit Resource">
                          <IconButton
                            size="small"
                            onClick={() => void openResourceEditor(row, "edit")}
                            aria-label="Edit resource"
                          >
                            <EditIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </TableCell>
                    </TableRow>
                  ))}
                  {filteredData.length === 0 && (
                    <TableRow>
                      <TableCell
                        colSpan={selectedResourceType.columns.length + 1}
                        align="center"
                        sx={{ py: 4 }}
                      >
                        <Typography variant="body2" color="text.secondary">
                          No {selectedResourceType.name.toLowerCase()} found.
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Box>
      </Paper>

      <Dialog
        open={editorOpen}
        onClose={closeResourceEditor}
        fullWidth
        maxWidth="lg"
        slotProps={{
          paper: {
            sx: {
              width: "min(92vw, 1100px)",
              height: "min(80vh, 820px)",
              maxHeight: "calc(100vh - 48px)",
              borderRadius: 2,
              overflow: "hidden",
            },
          },
        }}
      >
        <DialogTitle
          sx={{ px: 3, py: 2, borderBottom: 1, borderColor: "divider" }}
        >
          <Stack
            direction="row"
            justifyContent={"space-between" as const}
            alignItems={"center" as const}
            spacing={2}
          >
            <Box sx={{ minWidth: 0 }}>
              <Typography
                variant="h6"
                sx={{ fontSize: "1.1rem", fontWeight: 650 }}
              >
                {editorMode === "view" ? "View YAML" : "Edit Resource"}
              </Typography>
              <Typography variant="body2" color="text.secondary" noWrap>
                {editorTitle} · {selectedResourceType.name}
              </Typography>
            </Box>
            <IconButton
              onClick={closeResourceEditor}
              disabled={saving}
              aria-label="Close editor"
              size="small"
            >
              <CloseIcon />
            </IconButton>
          </Stack>
        </DialogTitle>
        <DialogContent
          dividers
          sx={{
            display: "flex",
            flexDirection: "column",
            minHeight: 0,
            p: 2,
            overflow: "hidden",
          }}
        >
          {editorError && (
            <Alert severity="error" sx={{ mb: 1 }}>
              {editorError}
            </Alert>
          )}
          {loadingEditor ? (
            <Stack
              sx={{ flex: 1, alignItems: "center", justifyContent: "center" }}
            >
              <CircularProgress />
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Loading resource...
              </Typography>
            </Stack>
          ) : (
            <Box
              sx={{
                flex: 1,
                minHeight: 0,
                height: "100%",
                border: 1,
                borderColor: "divider",
                borderRadius: 1,
                overflow: "hidden",
              }}
            >
              <Editor
                height="100%"
                language="yaml"
                theme="vs-dark"
                value={editorText}
                onChange={(value) => setEditorText(value ?? "")}
                options={{
                  readOnly: editorMode === "view" || saving,
                  minimap: { enabled: false },
                  wordWrap: "on",
                  automaticLayout: true,
                  scrollBeyondLastLine: false,
                  tabSize: 2,
                  formatOnPaste: true,
                  formatOnType: true,
                  padding: { top: 12 },
                }}
              />
            </Box>
          )}
        </DialogContent>
        <DialogActions
          sx={{ px: 3, py: 1.5, borderTop: 1, borderColor: "divider" }}
        >
          <Button onClick={closeResourceEditor} disabled={saving}>
            {editorMode === "view" ? "Close" : "Cancel"}
          </Button>
          {editorMode === "edit" && (
            <Button
              variant="contained"
              startIcon={<SaveIcon />}
              onClick={() => void saveResource()}
              disabled={
                saving ||
                loadingEditor ||
                !editorText ||
                editorText === originalText
              }
            >
              {saving ? "Saving..." : "Save"}
            </Button>
          )}
        </DialogActions>
      </Dialog>

      <Snackbar
        open={Boolean(notice)}
        autoHideDuration={3000}
        onClose={() => setNotice("")}
        message={notice}
      />
    </Box>
  );
};
