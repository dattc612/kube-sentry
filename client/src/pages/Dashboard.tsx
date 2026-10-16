import React, { useState } from "react";
import { Box, Tabs, Tab, ThemeProvider } from "@mui/material";
import {
  Dns as PodIcon,
  Storage as ResourceIcon,
  AccountTree as WorkflowIcon,
} from "@mui/icons-material";

import { theme } from "../theme";
import { PodOperationsView } from "../components/PodOperationsView";
import { K8sResourceManagement } from "../components/K8sResourceManagement";
import { GlobalAutomationCenter } from "../components/GlobalAutomationCenter";

export const Dashboard: React.FC = () => {
  const [globalView, setGlobalView] = useState<number>(0);

  return (
    <ThemeProvider theme={theme}>
      <Box
        sx={{
          p: 2,
          bgcolor: "background.default",
          height: "100vh",
          maxHeight: "100vh",
          width: "100vw",
          overflow: "hidden",
          boxSizing: "border-box",
          display: "flex",
          flexDirection: "column",
        }}
      >
        {/* Global View Selector Header */}
        <Box sx={{ mb: 1.5, flexShrink: 0 }}>
          <Tabs
            value={globalView}
            onChange={(_, val) => setGlobalView(val)}
            sx={{ minHeight: 36, "& .MuiTab-root": { minHeight: 36, py: 0 } }}
          >
            <Tab
              icon={<PodIcon fontSize="small" />}
              iconPosition="start"
              label="Pods & Containers"
            />
            <Tab
              icon={<ResourceIcon fontSize="small" />}
              iconPosition="start"
              label="K8s Resource Management"
            />
            <Tab
              icon={<WorkflowIcon fontSize="small" />}
              iconPosition="start"
              label="Automation Center"
            />
          </Tabs>
        </Box>

        {/* View Mode 0: Pod Operations */}
        {globalView === 0 && <PodOperationsView />}

        {/* View Mode 1: Unified K8s Resource Management (Deployments, Services, etc.) */}
        {globalView === 1 && <K8sResourceManagement />}

        {/* View Mode 2: Global Automation Center */}
        {globalView === 2 && (
          <Box
            sx={{
              flexGrow: 1,
              minHeight: 0,
              height: "100%",
              overflow: "hidden",
            }}
          >
            <GlobalAutomationCenter />
          </Box>
        )}
      </Box>
    </ThemeProvider>
  );
};

export default Dashboard;
