import React, { useState } from "react";
import { Box, Paper, Tabs, Tab, Stack, Typography, Chip } from "@mui/material";
import {
  ListAlt as CatalogueIcon,
  History as HistoryIcon,
  CheckCircle,
  PlayCircleFilled,
  Error as ErrorIcon,
} from "@mui/icons-material";
import { AvailableAutomationsList } from "./AvailableAutomationsList";
import { AutomationTrackerTab } from "./AutomationTrackerTab";

export const GlobalAutomationCenter: React.FC = () => {
  const [subTab, setSubTab] = useState<number>(0);

  return (
    <Paper
      elevation={0}
      variant="outlined"
      sx={{
        p: 2.5,
        height: "100%",
        display: "flex",
        flexDirection: "column",
        boxSizing: "border-box",
        overflow: "hidden",
      }}
    >
      {/* Header Stat Summary Bar */}
      <Stack
        direction="row"
        spacing={3}
        sx={{ mb: 2, flexShrink: 0, alignItems: "center" }}
      >
        <Box>
          <Typography variant="h6" sx={{ fontWeight: 700 }}>
            StackStorm Automation Engine
          </Typography>
          <Typography variant="caption" color="text.secondary">
            Cluster Event Triggers & Rule Executions
          </Typography>
        </Box>

        <Stack direction="row" spacing={1.5} sx={{ ml: "auto !important" }}>
          <Chip
            icon={<CheckCircle />}
            label="? Succeeded"
            color="success"
            variant="outlined"
            size="small"
          />
          <Chip
            icon={<PlayCircleFilled />}
            label="? Active"
            color="warning"
            variant="outlined"
            size="small"
          />
          <Chip
            icon={<ErrorIcon />}
            label="? Failed"
            color="error"
            variant="outlined"
            size="small"
          />
        </Stack>
      </Stack>

      {/* Tabs */}
      <Box
        sx={{ borderBottom: 1, borderColor: "divider", mb: 2, flexShrink: 0 }}
      >
        <Tabs value={subTab} onChange={(_, val) => setSubTab(val)}>
          <Tab
            icon={<CatalogueIcon fontSize="small" />}
            iconPosition="start"
            label="Available Actions & Rules"
          />
          <Tab
            icon={<HistoryIcon fontSize="small" />}
            iconPosition="start"
            label="Global Execution History"
          />
        </Tabs>
      </Box>

      {/* Tab Panels */}
      <Box sx={{ flexGrow: 1, minHeight: 0, overflow: "hidden" }}>
        {subTab === 0 && <AvailableAutomationsList />}
        {subTab === 1 && <AutomationTrackerTab />}
      </Box>
    </Paper>
  );
};
