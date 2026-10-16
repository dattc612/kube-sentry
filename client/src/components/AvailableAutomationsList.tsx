import React, { useState, useEffect, useCallback } from "react";
import {
  Box,
  Typography,
  TextField,
  InputAdornment,
  Grid,
  Card,
  CardContent,
  Chip,
  Button,
  Stack,
  Divider,
  CircularProgress,
  Alert,
  Snackbar,
} from "@mui/material";
import {
  Search as SearchIcon,
  PlayArrow as RunIcon,
  AutoFixHigh as AutoIcon,
  Code as CodeIcon,
  Notifications as NotifyIcon,
  Refresh as RefreshIcon,
} from "@mui/icons-material";
import { darkScrollbarStyles } from "../theme";
import { automationApi } from "../api/automationApi";
import { Automation } from "../models/automation";
import { TriggerAutomationDialog } from "./TriggerAutomationDialog";

export const AvailableAutomationsList: React.FC = () => {
  const [selectedAutomation, setSelectedAutomation] =
    useState<Automation | null>(null);

  const [triggerDialogOpen, setTriggerDialogOpen] = useState(false);

  const [automations, setAutomations] = useState<Automation[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const [query, setQuery] = useState<string>("");
  const [selectedPack, setSelectedPack] = useState<string>("all");

  // State to track job triggering and user notifications
  const [triggeringCode, setTriggeringCode] = useState<string | null>(null);
  const [snackbarMessage, setSnackbarMessage] = useState<string | null>(null);

  // Fetch automations from backend API
  const fetchAutomations = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await automationApi.getAutomations();
      setAutomations(data.data);
    } catch (err: any) {
      setError(
        err.response?.data?.error ||
          err.message ||
          "Failed to load automations list from API",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAutomations();
  }, [fetchAutomations]);

  const availablePacks = [
    "all",
    ...Array.from(
      new Set(automations.map((item) => item.pack).filter(Boolean)),
    ),
  ];

  const filteredActions = automations.filter((item) => {
    const matchesPack = selectedPack === "all" || item.pack === selectedPack;
    const matchesQuery =
      item.name.toLowerCase().includes(query.toLowerCase()) ||
      item.description?.toLowerCase().includes(query.toLowerCase()) ||
      item.code?.toLowerCase().includes(query.toLowerCase()) ||
      item.pack?.toLowerCase().includes(query.toLowerCase());

    return matchesPack && matchesQuery;
  });

  // Handle triggering a job execution
  const handleTrigger = async (
    automationCode: string,
    input: Record<string, unknown>,
  ) => {
    setTriggeringCode(automationCode);

    try {
      const res = await automationApi.triggerJob(automationCode, input);

      setSnackbarMessage(
        `Job queued successfully! (Code: ${res.execution_code})`,
      );

      setTriggerDialogOpen(false);
      setSelectedAutomation(null);
    } catch (err: any) {
      setSnackbarMessage(
        `Failed to trigger job: ${
          err.response?.data?.error || err.message || "Unknown error"
        }`,
      );
    } finally {
      setTriggeringCode(null);
    }
  };

  const getBadgeIcon = (type: Automation["type"]) => {
    switch (type) {
      case "rule":
        return <AutoIcon fontSize="small" />;
      case "workflow":
        return <CodeIcon fontSize="small" />;
      default:
        return <NotifyIcon fontSize="small" />;
    }
  };

  return (
    <Box
      sx={{
        height: "100%",
        display: "flex",
        flexDirection: "column",
        minHeight: 0,
      }}
    >
      {/* Search Bar, Pack Filters & Refresh */}
      <Stack
        direction="row"
        spacing={2}
        sx={{ mb: 2, flexShrink: 0, alignItems: "center" }}
      >
        <TextField
          size="small"
          placeholder="Search actions, codes, packs..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          sx={{ maxWidth: 360 }}
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

        <Stack direction="row" spacing={1} sx={{ overflowX: "auto", py: 0.5 }}>
          {availablePacks.map((pack) => (
            <Chip
              key={pack}
              label={pack}
              clickable
              color={selectedPack === pack ? "primary" : "default"}
              variant={selectedPack === pack ? "filled" : "outlined"}
              size="small"
              onClick={() => setSelectedPack(pack)}
            />
          ))}
        </Stack>

        <Button
          size="small"
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={fetchAutomations}
          disabled={loading}
          sx={{ ml: "auto" }}
        >
          Refresh
        </Button>
      </Stack>

      {/* Scrollable Action Cards Grid */}
      <Box
        sx={{
          flexGrow: 1,
          minHeight: 0,
          overflow: "auto",
          pr: 1,
          ...darkScrollbarStyles,
        }}
      >
        {loading ? (
          <Box
            sx={{
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              height: 240,
            }}
          >
            <CircularProgress />
          </Box>
        ) : error ? (
          <Alert severity="error" sx={{ my: 2 }}>
            {error}
          </Alert>
        ) : filteredActions.length === 0 ? (
          <Box sx={{ textAlign: "center", py: 6, color: "text.secondary" }}>
            <Typography variant="body1">No automations found.</Typography>
          </Box>
        ) : (
          <Grid container spacing={2}>
            {filteredActions.map((item) => {
              const isActive = item.status === "active";
              const isTriggering = triggeringCode === item.code;

              return (
                <Grid size={{ xs: 12, md: 6, lg: 4 }} key={item.id}>
                  <Card
                    variant="outlined"
                    sx={{
                      bgcolor: "background.paper",
                      height: "100%",
                      display: "flex",
                      flexDirection: "column",
                    }}
                  >
                    <CardContent sx={{ flexGrow: 1, p: 2 }}>
                      <Stack
                        direction="row"
                        sx={{
                          justifyContent: "space-between",
                          alignItems: "flex-start",
                          mb: 1,
                        }}
                      >
                        <Chip
                          icon={getBadgeIcon(item.type)}
                          label={`${item.pack || "default"}.${item.type}`}
                          size="small"
                          variant="outlined"
                          color="primary"
                        />
                        <Chip
                          label={isActive ? "Active" : "Disabled"}
                          size="small"
                          color={isActive ? "success" : "default"}
                          sx={{ height: 20, fontSize: "0.7rem" }}
                        />
                      </Stack>

                      <Typography
                        variant="subtitle1"
                        sx={{
                          fontWeight: 600,
                          fontFamily: "monospace",
                          mb: 0.5,
                        }}
                      >
                        {item.name}
                      </Typography>

                      <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{ mb: 2, minHeight: 40 }}
                      >
                        {item.description || "No description provided."}
                      </Typography>

                      <Divider sx={{ mb: 1.5 }} />

                      <Stack
                        direction="row"
                        sx={{
                          justifyContent: "space-between",
                          alignItems: "center",
                        }}
                      >
                        <Typography
                          variant="caption"
                          color="text.secondary"
                          sx={{ fontFamily: "monospace" }}
                        >
                          Code: {item.code}
                        </Typography>
                        <Button
                          size="small"
                          variant="contained"
                          startIcon={
                            isTriggering ? (
                              <CircularProgress size={16} color="inherit" />
                            ) : (
                              <RunIcon />
                            )
                          }
                          disabled={!isActive || isTriggering}
                          onClick={() => {
                            setSelectedAutomation(item);
                            setTriggerDialogOpen(true);
                          }}
                        >
                          {isTriggering ? "Queuing..." : "Trigger"}
                        </Button>
                      </Stack>
                    </CardContent>
                  </Card>
                </Grid>
              );
            })}
          </Grid>
        )}
      </Box>

      {/* Notification Toast */}
      <Snackbar
        open={Boolean(snackbarMessage)}
        autoHideDuration={4000}
        onClose={() => setSnackbarMessage(null)}
        message={snackbarMessage}
      />
      <TriggerAutomationDialog
        open={triggerDialogOpen}
        automation={selectedAutomation}
        loading={
          selectedAutomation
            ? triggeringCode === selectedAutomation.code
            : false
        }
        onClose={() => {
          if (!triggeringCode) {
            setTriggerDialogOpen(false);
            setSelectedAutomation(null);
          }
        }}
        onTrigger={async (input) => {
          if (!selectedAutomation) {
            return;
          }

          await handleTrigger(selectedAutomation.code, input);
        }}
      />
    </Box>
  );
};
