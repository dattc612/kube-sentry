import React from "react";
import {
  Paper,
  Box,
  Stack,
  IconButton,
  Alert,
  TextField,
  InputAdornment,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  CircularProgress,
  Divider,
  List,
  ListItemButton,
  Typography,
  Chip,
  SelectChangeEvent,
} from "@mui/material";
import {
  Refresh as RefreshIcon,
  Search as SearchIcon,
  CheckCircle as RunningIcon,
  Error as ErrorIcon,
  Pending as PendingIcon,
} from "@mui/icons-material";
import { darkScrollbarStyles } from "../theme";
import { PodItem, PodStatus } from "../models/k8sResource";

interface PodListSidebarProps {
  pods: PodItem[];
  selectedPod: PodItem | null;
  namespaces: { name: string }[];
  namespaceFilter: string;
  searchQuery: string;
  loadingNs: boolean;
  loadingPods: boolean;
  displayError: string | null;
  onSearchChange: (value: string) => void;
  onNamespaceChange: (event: SelectChangeEvent) => void;
  onSelectPod: (pod: PodItem) => void;
  onRefresh: () => void;
}

export const PodListSidebar: React.FC<PodListSidebarProps> = ({
  pods,
  selectedPod,
  namespaces,
  namespaceFilter,
  searchQuery,
  loadingNs,
  loadingPods,
  displayError,
  onSearchChange,
  onNamespaceChange,
  onSelectPod,
  onRefresh,
}) => {
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
    <Paper
      elevation={0}
      variant="outlined"
      sx={{
        p: 2,
        height: "100%",
        display: "flex",
        flexDirection: "column",
        boxSizing: "border-box",
        borderColor: "divider",
        overflow: "hidden",
      }}
    >
      <Box sx={{ flexShrink: 0 }}>
        <Stack
          direction="row"
          sx={{
            justifyContent: "space-between",
            alignItems: "center",
            mb: 1.5,
          }}
        >
          <Typography
            variant="overline"
            sx={{
              fontWeight: 700,
              letterSpacing: 1.2,
              color: "text.secondary",
            }}
          >
            PACKAGES & PODS
          </Typography>
          <IconButton
            size="small"
            aria-label="refresh list"
            onClick={onRefresh}
          >
            <RefreshIcon fontSize="small" />
          </IconButton>
        </Stack>

        {displayError && (
          <Alert severity="error" sx={{ mb: 2, py: 0 }}>
            {displayError}
          </Alert>
        )}

        <Stack direction="row" spacing={1} sx={{ mb: 2 }}>
          <TextField
            size="small"
            placeholder="Filter execution targets..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            fullWidth
            slotProps={{
              input: {
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon fontSize="small" />
                  </InputAdornment>
                ),
              },
            }}
          />
          <FormControl size="small" sx={{ minWidth: 130 }}>
            <InputLabel id="namespace-select-label">Namespace</InputLabel>
            <Select
              labelId="namespace-select-label"
              value={namespaceFilter}
              label="Namespace"
              onChange={onNamespaceChange}
              disabled={loadingNs}
            >
              <MenuItem value="all">All</MenuItem>
              {namespaces.map((ns) => (
                <MenuItem key={ns.name} value={ns.name}>
                  {ns.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Stack>
        <Divider sx={{ mb: 1 }} />
      </Box>

      {/* Scrollable Pod List */}
      <List
        sx={{
          flexGrow: 1,
          minHeight: 0,
          overflow: "auto",
          pr: 0.5,
          ...darkScrollbarStyles,
        }}
      >
        {loadingPods ? (
          <Stack sx={{ alignItems: "center", justifyContent: "center", py: 4 }}>
            <CircularProgress size={24} />
            <Typography variant="caption" sx={{ mt: 1 }} color="text.secondary">
              Querying cluster status...
            </Typography>
          </Stack>
        ) : pods.length > 0 ? (
          pods.map((pod) => (
            <ListItemButton
              key={pod.id}
              selected={selectedPod?.id === pod.id}
              onClick={() => onSelectPod(pod)}
              sx={{
                borderRadius: 1,
                mb: 0.8,
                flexDirection: "column",
                alignItems: "flex-start",
                borderLeft:
                  selectedPod?.id === pod.id
                    ? "3px solid #00b4d8"
                    : "3px solid transparent",
                bgcolor:
                  selectedPod?.id === pod.id
                    ? "rgba(0, 180, 216, 0.08)"
                    : "transparent",
              }}
            >
              <Stack
                direction="row"
                sx={{ justifyContent: "space-between", width: "100%", mb: 0.5 }}
              >
                <Typography
                  variant="body2"
                  noWrap
                  sx={{ fontWeight: 600, maxWidth: "65%" }}
                >
                  {pod.name}
                </Typography>
                {getStatusChip(pod.status)}
              </Stack>
              <Typography variant="caption" color="text.secondary">
                {pod.namespace} • Age: {pod.age}
              </Typography>
            </ListItemButton>
          ))
        ) : (
          <Typography
            variant="caption"
            sx={{ p: 2, display: "block", color: "text.secondary" }}
          >
            No target pods matched.
          </Typography>
        )}
      </List>
    </Paper>
  );
};
