import React, { useEffect, useRef, useState } from "react";
import {
  Box,
  CircularProgress,
  Stack,
  Typography,
  IconButton,
  Tooltip,
} from "@mui/material";
import {
  ArrowDownward as ScrollBottomIcon,
  VerticalAlignBottom as AutoScrollIcon,
} from "@mui/icons-material";
import { darkScrollbarStyles } from "../theme";

interface ConsoleLogsTabProps {
  logs: string | null;
  logsLoading: boolean;
}

export const ConsoleLogsTab: React.FC<ConsoleLogsTabProps> = ({
  logs,
  logsLoading,
}) => {
  const logContainerRef = useRef<HTMLPreElement | null>(null);
  const [autoScroll, setAutoScroll] = useState<boolean>(true);

  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const scrollToBottom = () => {
    if (logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  };

  return (
    <Box
      sx={{
        position: "relative",
        height: "100%",
        width: "100%",
        display: "flex",
        flexDirection: "column",
        minHeight: 0,
      }}
    >
      <Stack
        direction="row"
        spacing={0.5}
        sx={{
          position: "absolute",
          top: 10,
          right: 14,
          zIndex: 10,
          bgcolor: "rgba(13, 17, 23, 0.85)",
          backdropFilter: "blur(4px)",
          borderRadius: 1,
          p: 0.5,
          border: "1px solid rgba(255, 255, 255, 0.1)",
        }}
      >
        <Tooltip
          title={autoScroll ? "Disable Auto-scroll" : "Enable Auto-scroll"}
        >
          <IconButton
            size="small"
            color={autoScroll ? "primary" : "default"}
            onClick={() => setAutoScroll(!autoScroll)}
          >
            <AutoScrollIcon fontSize="small" />
          </IconButton>
        </Tooltip>
        <Tooltip title="Scroll to Bottom">
          <IconButton size="small" onClick={scrollToBottom}>
            <ScrollBottomIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      </Stack>

      {/* Internal Scrollable Terminal Box */}
      <Box
        ref={logContainerRef}
        component="pre"
        sx={{
          p: 2,
          m: 0,
          flexGrow: 1,
          minHeight: 0, // Critical to prevent expanding outer window
          overflow: "auto", // 👈 Enables independent vertical/horizontal scrolling
          borderRadius: 1,
          bgcolor: "#0d1117",
          color: "#c9d1d9",
          fontFamily: "'Fira Code', 'Roboto Mono', monospace",
          fontSize: "0.8rem",
          lineHeight: 1.5,
          border: "1px solid",
          borderColor: "divider",
          whiteSpace: "pre-wrap",
          wordBreak: "break-all",
          boxSizing: "border-box",
          ...darkScrollbarStyles,
        }}
      >
        {logsLoading ? (
          <Stack
            direction="row"
            spacing={1}
            sx={{ alignItems: "center", py: 2 }}
          >
            <CircularProgress size={16} color="inherit" />
            <Typography variant="caption">
              Fetching console streams...
            </Typography>
          </Stack>
        ) : (
          <code>
            {logs || "No console logs streams present for this target."}
          </code>
        )}
      </Box>
    </Box>
  );
};
