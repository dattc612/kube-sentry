import React, { useState } from "react";
import {
  Box,
  Paper,
  Typography,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  TextField,
  InputAdornment,
  IconButton,
  Tooltip,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  ListSubheader,
  Divider,
  CircularProgress,
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
  Code as YamlIcon,
  DeleteOutlined as DeleteIcon,
  Edit as EditIcon,
} from "@mui/icons-material";

import { useK8sResources } from "../hooks/useK8sResources";

interface ResourceItem {
  id: string;
  name: string;
  label: string;
  icon: React.ReactNode;
  category: string;
  columns: string[];
}

const RESOURCE_TYPES: ResourceItem[] = [
  // 1. Workloads
  {
    id: "pods",
    name: "Pods",
    label: "Pods",
    icon: <PodIcon fontSize="small" />,
    category: "Workloads",
    columns: ["Name", "Namespace", "Ready", "Status", "Restarts", "Age"],
  },
  {
    id: "deployments",
    name: "Deployments",
    label: "Deployments",
    icon: <DeploymentIcon fontSize="small" />,
    category: "Workloads",
    columns: ["Name", "Namespace", "Ready", "Up-to-date", "Available", "Age"],
  },
  {
    id: "statefulsets",
    name: "StatefulSets",
    label: "StatefulSets",
    icon: <DeploymentIcon fontSize="small" />,
    category: "Workloads",
    columns: ["Name", "Namespace", "Ready", "Age"],
  },
  {
    id: "daemonsets",
    name: "DaemonSets",
    label: "DaemonSets",
    icon: <DeploymentIcon fontSize="small" />,
    category: "Workloads",
    columns: ["Name", "Namespace", "Desired", "Current", "Ready", "Age"],
  },
  {
    id: "cronjobs",
    name: "CronJobs",
    label: "CronJobs",
    icon: <CronJobIcon fontSize="small" />,
    category: "Workloads",
    columns: ["Name", "Namespace", "Schedule", "Suspend", "Active", "Age"],
  },

  // 2. Network
  {
    id: "services",
    name: "Services",
    label: "Services",
    icon: <ServiceIcon fontSize="small" />,
    category: "Network",
    columns: [
      "Name",
      "Namespace",
      "Type",
      "Cluster IP",
      "External IP",
      "Port(s)",
    ],
  },
  {
    id: "ingresses",
    name: "Ingresses",
    label: "Ingresses",
    icon: <IngressIcon fontSize="small" />,
    category: "Network",
    columns: ["Name", "Namespace", "Class", "Hosts", "Address", "Ports"],
  },

  // 3. Storage
  {
    id: "pvcs",
    name: "PersistentVolumeClaims",
    label: "PVCs",
    icon: <VolumeIcon fontSize="small" />,
    category: "Storage",
    columns: [
      "Name",
      "Namespace",
      "Status",
      "Volume",
      "Capacity",
      "StorageClass",
    ],
  },
  {
    id: "pvs",
    name: "PersistentVolumes",
    label: "PVs",
    icon: <StorageIcon fontSize="small" />,
    category: "Storage",
    columns: [
      "Name",
      "Capacity",
      "Access Modes",
      "Reclaim Policy",
      "Status",
      "StorageClass",
    ],
  },

  // 4. Config & Secrets
  {
    id: "configmaps",
    name: "ConfigMaps",
    label: "ConfigMaps",
    icon: <ConfigMapIcon fontSize="small" />,
    category: "Config",
    columns: ["Name", "Namespace", "Data Keys", "Age"],
  },
  {
    id: "secrets",
    name: "Secrets",
    label: "Secrets",
    icon: <SecretIcon fontSize="small" />,
    category: "Config",
    columns: ["Name", "Namespace", "Type", "Data Keys", "Age"],
  },

  // 5. Access & RBAC
  {
    id: "serviceaccounts",
    name: "ServiceAccounts",
    label: "ServiceAccounts",
    icon: <RbacIcon fontSize="small" />,
    category: "Access Control",
    columns: ["Name", "Namespace", "Secrets", "Age"],
  },
  {
    id: "roles",
    name: "Roles & ClusterRoles",
    label: "Roles",
    icon: <RbacIcon fontSize="small" />,
    category: "Access Control",
    columns: ["Name", "Namespace", "Kind", "Age"],
  },

  // 6. Cluster & Admin
  {
    id: "nodes",
    name: "Nodes",
    label: "Nodes",
    icon: <NodeIcon fontSize="small" />,
    category: "Cluster",
    columns: [
      "Name",
      "Status",
      "Roles",
      "Version",
      "Internal IP",
      "CPU %",
      "Mem %",
    ],
  },
  {
    id: "namespaces",
    name: "Namespaces",
    label: "Namespaces",
    icon: <NamespaceIcon fontSize="small" />,
    category: "Cluster",
    columns: ["Name", "Status", "Age"],
  },
];

export const K8sResourceManagement: React.FC = () => {
  const [selectedResourceType, setSelectedResourceType] =
    useState<ResourceItem>(RESOURCE_TYPES[0]);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [namespaceFilter, setNamespaceFilter] = useState<string>("all");

  // Lazy-load data using your custom hook bound to the active tab & namespace
  const {
    data: rawData,
    loading,
    error,
    refetch,
  } = useK8sResources(selectedResourceType.id, namespaceFilter);

  const categories = Array.from(new Set(RESOURCE_TYPES.map((r) => r.category)));

  // Filter raw API data based on search query
  const filteredData = (rawData || []).filter((item) => {
    return item.name?.toLowerCase().includes(searchQuery.toLowerCase());
  });

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
      {/* Left Resource Category Navigation Sidebar */}
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
                    setSearchQuery(""); // reset search query on tab change
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
                    primaryTypographyProps={{
                      fontSize: "0.85rem",
                      fontWeight:
                        selectedResourceType.id === res.id ? 600 : 400,
                    }}
                  />
                </ListItemButton>
              ))}
              <Divider sx={{ my: 1 }} />
            </React.Fragment>
          ))}
        </List>
      </Paper>

      {/* Main Content Area: Resource View Table */}
      <Paper
        elevation={0}
        variant="outlined"
        sx={{
          flexGrow: 1,
          height: "100%",
          display: "flex",
          flexDirection: "column",
          p: 2,
          boxSizing: "border-box",
          overflow: "hidden",
        }}
      >
        {/* Header Title Bar */}
        <Stack
          direction="row"
          sx={{ justifyContent: "space-between", alignItems: "center", mb: 2 }}
        >
          <Stack direction="row" spacing={1} sx={{ alignItems: "center" }}>
            {selectedResourceType.icon}
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              {selectedResourceType.name} ({filteredData.length})
            </Typography>
          </Stack>
          <Tooltip title="Refresh Cluster Resources">
            <IconButton size="small" onClick={refetch} disabled={loading}>
              <RefreshIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        </Stack>

        {/* Search & Namespace Filter Toolbar */}
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
              <MenuItem value="all">All Namespaces</MenuItem>
              <MenuItem value="default">default</MenuItem>
              <MenuItem value="kube-system">kube-system</MenuItem>
              <MenuItem value="finance">finance</MenuItem>
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

        {/* Dynamic Table Content with Loading & Error States */}
        <Box
          sx={{
            flexGrow: 1,
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
                    {selectedResourceType.columns.map((col, idx) => (
                      <TableCell key={idx}>{col}</TableCell>
                    ))}
                    <TableCell align="right">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {filteredData.map((row, rIdx) => (
                    <TableRow key={rIdx} hover>
                      <TableCell sx={{ fontWeight: 500 }}>{row.name}</TableCell>
                      {row.namespace && (
                        <TableCell>
                          <Chip
                            label={row.namespace}
                            size="small"
                            variant="outlined"
                            sx={{ borderRadius: "4px" }}
                          />
                        </TableCell>
                      )}
                      <TableCell>
                        {row.status ||
                          row.ready ||
                          row.type ||
                          row.capacity ||
                          "Active"}
                      </TableCell>
                      <TableCell>{row.age || row.version || "N/A"}</TableCell>

                      <TableCell align="right">
                        <Tooltip title="View YAML">
                          <IconButton size="small">
                            <YamlIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Edit Resource">
                          <IconButton size="small">
                            <EditIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Delete">
                          <IconButton size="small" color="error">
                            <DeleteIcon fontSize="small" />
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
    </Box>
  );
};
