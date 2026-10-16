import React, { useCallback, useEffect, useState } from "react";
import {
  Box,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TableSortLabel,
  TablePagination,
  Chip,
  IconButton,
  Typography,
  Collapse,
  Paper,
  Button,
  Stack,
  CircularProgress,
  Tooltip,
} from "@mui/material";
import {
  KeyboardArrowDown,
  KeyboardArrowUp,
  Refresh as RefreshIcon,
  CheckCircle,
  Error as ErrorIcon,
  DataObject as PayloadIcon,
  ContentCopy as CopyIcon,
  Check as CheckIcon,
} from "@mui/icons-material";
import { darkScrollbarStyles } from "../theme";
import { automationApi } from "../api/automationApi";
import type { Execution, ExecutionStep } from "../models/automation";

// Sorting type definitions
type Order = "asc" | "desc";
type SortableColumn =
  | "automation_code"
  | "status"
  | "created_at"
  | "duration_ms";

const formatPrettyJson = (input: any): string => {
  if (!input) return "";
  try {
    const parsed = typeof input === "string" ? JSON.parse(input) : input;
    return JSON.stringify(parsed, null, 2);
  } catch (e) {
    return String(input);
  }
};

const getStatusChip = (status?: string) => {
  const normalizedStatus = status?.toLowerCase();

  switch (normalizedStatus) {
    case "succeeded":
    case "success":
      return (
        <Chip
          icon={<CheckCircle fontSize="small" />}
          label="SUCCEEDED"
          color="success"
          size="small"
        />
      );

    case "running":
    case "in_progress":
      return (
        <Chip
          icon={<CircularProgress size={14} color="inherit" sx={{ ml: 0.5 }} />}
          label="RUNNING"
          color="warning"
          size="small"
        />
      );

    case "failed":
    case "error":
      return (
        <Chip
          icon={<ErrorIcon fontSize="small" />}
          label="FAILED"
          color="error"
          size="small"
        />
      );

    default:
      return (
        <Chip
          label={status?.toUpperCase() || "INFO"}
          size="small"
          variant="outlined"
        />
      );
  }
};

const getStatusColor = (status?: string) => {
  switch (status?.toLowerCase()) {
    case "succeeded":
    case "success":
      return "#238636";
    case "running":
    case "in_progress":
      return "#d29922";
    case "failed":
    case "error":
      return "#f85149";
    default:
      return "#8b949e";
  }
};
const JsonTooltipContent: React.FC<{ input: any }> = ({ input }) => {
  const [copied, setCopied] = useState(false);
  const formattedJson = formatPrettyJson(input);

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigator.clipboard.writeText(formattedJson);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Box sx={{ p: 0.5 }}>
      <Stack
        direction="row"
        sx={{
          justifyContent: "space-between",
          alignItems: "center",
          borderBottom: "1px solid #30363d",
          pb: 0.5,
          mb: 0.5,
        }}
      >
        <Typography
          variant="caption"
          sx={{ color: "#8b949e", fontWeight: 700, fontFamily: "monospace" }}
        >
          JSON PAYLOAD
        </Typography>

        <Tooltip title={copied ? "Copied!" : "Copy JSON"} placement="top">
          <IconButton
            size="small"
            onClick={handleCopy}
            sx={{
              color: copied ? "#3fb950" : "#8b949e",
              p: "2px",
              "&:hover": { color: "#c9d1d9", bgcolor: "rgba(255,255,255,0.1)" },
            }}
          >
            {copied ? (
              <CheckIcon sx={{ fontSize: 14 }} />
            ) : (
              <CopyIcon sx={{ fontSize: 14 }} />
            )}
          </IconButton>
        </Tooltip>
      </Stack>

      <Box
        component="pre"
        sx={{
          m: 0,
          p: 0.5,
          maxHeight: 250,
          maxWidth: 600,
          overflow: "auto",
          whiteSpace: "pre-wrap",
          wordBreak: "break-all",
          fontSize: "0.75rem",
          fontFamily: "monospace",
          color: "#c9d1d9",
        }}
      >
        {formattedJson}
      </Box>
    </Box>
  );
};

const TaskRow: React.FC<{ row: Execution }> = ({ row }) => {
  const [open, setOpen] = useState(false);
  const [executionSteps, setExecutionSteps] = useState<ExecutionStep[]>([]);
  const [loadingLogs, setLoadingLogs] = useState(false);

  const fetchSteps = useCallback(async () => {
    setLoadingLogs(true);
    try {
      const steps = await automationApi.getExecutionSteps(Number(row.id));
      setExecutionSteps(steps ?? []);
    } catch (error) {
      console.error("Failed to load execution logs", error);
    } finally {
      setLoadingLogs(false);
    }
  }, [row.id]);

  const handleOpen = async () => {
    const nextOpen = !open;
    setOpen(nextOpen);

    if (nextOpen && executionSteps.length === 0) {
      await fetchSteps();
    }
  };

  return (
    <>
      <TableRow sx={{ "& > *": { borderBottom: "unset" } }}>
        <TableCell>
          <IconButton size="small" onClick={handleOpen}>
            {open ? <KeyboardArrowUp /> : <KeyboardArrowDown />}
          </IconButton>
        </TableCell>

        <TableCell component="th" scope="row" sx={{ fontFamily: "monospace" }}>
          {row.execution_code}
        </TableCell>
        <TableCell>{row.automation_code}</TableCell>
        <TableCell>{row.trigger_source}</TableCell>
        <TableCell align="center" sx={{ width: "60px" }}>
          {row.input ? (
            <Tooltip
              title={<JsonTooltipContent input={row.input} />}
              enterDelay={200}
              leaveDelay={300}
              slotProps={{
                tooltip: {
                  sx: {
                    maxWidth: 650,
                    width: "max-content",
                    bgcolor: "#0d1117",
                    border: "1px solid #30363d",
                    p: 1,
                  },
                },
                popper: {
                  sx: { pointerEvents: "auto" },
                },
              }}
            >
              <IconButton
                size="small"
                sx={{
                  color: "#58a6ff",
                  bgcolor: "rgba(56, 139, 253, 0.1)",
                  border: "1px solid rgba(56, 139, 253, 0.3)",
                  borderRadius: 1,
                  p: "2px",
                  "&:hover": {
                    bgcolor: "rgba(56, 139, 253, 0.25)",
                    borderColor: "#58a6ff",
                  },
                }}
              >
                <PayloadIcon sx={{ fontSize: 16 }} />
              </IconButton>
            </Tooltip>
          ) : (
            <Typography variant="caption" sx={{ color: "text.disabled" }}>
              -
            </Typography>
          )}
        </TableCell>
        <TableCell>{getStatusChip(row.status)}</TableCell>
        <TableCell>{row.created_at}</TableCell>
        <TableCell>{row.finished_at || "-"}</TableCell>
        <TableCell>{row.duration_ms ? `${row.duration_ms}ms` : "-"}</TableCell>
      </TableRow>

      <TableRow>
        <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={8}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box
              sx={{
                margin: 1.5,
                p: 2,
                bgcolor: "#0d1117",
                borderRadius: 1.5,
                border: "1px solid #30363d",
              }}
            >
              <Stack
                direction="row"
                sx={{
                  justifyContent: "space-between",
                  alignItems: "center",
                  mb: 1.5,
                }}
              >
                <Typography
                  variant="caption"
                  sx={{
                    fontWeight: 700,
                    color: "#58a6ff",
                    letterSpacing: "0.5px",
                  }}
                >
                  EXECUTION STEPS & LOG DETAILS
                </Typography>

                {/* 🔄 Per-Execution Step Refresh Button */}
                <Tooltip title="Refresh execution steps">
                  <span>
                    <IconButton
                      size="small"
                      onClick={fetchSteps}
                      disabled={loadingLogs}
                      sx={{ color: "#8b949e", "&:hover": { color: "#c9d1d9" } }}
                    >
                      {loadingLogs ? (
                        <CircularProgress size={14} color="inherit" />
                      ) : (
                        <RefreshIcon sx={{ fontSize: 16 }} />
                      )}
                    </IconButton>
                  </span>
                </Tooltip>
              </Stack>

              {loadingLogs && executionSteps.length === 0 ? (
                <Stack
                  direction="row"
                  spacing={1}
                  sx={{ py: 1, alignItems: "center" }}
                >
                  <CircularProgress size={16} />
                  <Typography variant="body2" color="text.secondary">
                    Loading execution steps...
                  </Typography>
                </Stack>
              ) : executionSteps.length > 0 ? (
                <Box sx={{ position: "relative", pl: 3, my: 1 }}>
                  <Box
                    sx={{
                      position: "absolute",
                      left: 11,
                      top: 14,
                      bottom: 14,
                      width: "2px",
                      bgcolor: "#30363d",
                      zIndex: 0,
                    }}
                  />

                  {executionSteps.map((step: ExecutionStep, index: number) => {
                    const dotColor = getStatusColor(step.status);
                    const timestamp =
                      step.log?.created_at ||
                      step.finished_at ||
                      step.started_at;

                    return (
                      <Box
                        key={step.id ?? index}
                        sx={{
                          position: "relative",
                          py: 1,
                          px: 1.5,
                          mb: 0.5,
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "space-between",
                          borderRadius: 1,
                          transition: "background-color 0.15s ease",
                          "&:hover": {
                            bgcolor: "rgba(255, 255, 255, 0.04)",
                          },
                        }}
                      >
                        <Box
                          sx={{
                            position: "absolute",
                            left: -17,
                            width: 10,
                            height: 10,
                            borderRadius: "50%",
                            bgcolor: dotColor,
                            boxShadow: `0 0 0 3px #0d1117, 0 0 6px ${dotColor}`,
                            zIndex: 1,
                          }}
                        />

                        <Stack
                          direction="row"
                          spacing={1.5}
                          sx={{ alignItems: "center" }}
                        >
                          {/* {getStatusChip(step.status || "succeeded")} */}

                          <Typography
                            variant="body2"
                            sx={{
                              fontFamily: "monospace",
                              color: "#c9d1d9",
                              fontSize: "0.85rem",
                              whiteSpace: "pre-wrap",
                            }}
                            dangerouslySetInnerHTML={{
                              __html:
                                step.log?.message ||
                                step.name ||
                                `Step #${index + 1}`,
                            }}
                          />
                        </Stack>

                        {timestamp && (
                          <Typography
                            variant="caption"
                            sx={{
                              fontFamily: "monospace",
                              color: "#8b949e",
                              fontSize: "0.75rem",
                              flexShrink: 0,
                              ml: 2,
                            }}
                          >
                            {timestamp}
                          </Typography>
                        )}
                      </Box>
                    );
                  })}
                </Box>
              ) : (
                <Typography variant="body2" sx={{ color: "#8b949e", py: 0.5 }}>
                  No execution steps recorded.
                </Typography>
              )}
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
};

export const AutomationTrackerTab: React.FC = () => {
  const [tasks, setTasks] = useState<Execution[]>([]);
  const [loading, setLoading] = useState(false);

  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(20);
  const [totalCount, setTotalCount] = useState(0);

  const [orderBy, setOrderBy] = useState<SortableColumn>("created_at");
  const [order, setOrder] = useState<Order>("desc");

  const loadExecutions = useCallback(async () => {
    setLoading(true);

    try {
      const offset = page * rowsPerPage;

      const response = await automationApi.listExecutions(
        rowsPerPage,
        offset,
        orderBy,
        order,
      );

      const executions = response.data || [];
      setTasks(executions);

      const total =
        response.total ??
        (offset + executions.length < rowsPerPage
          ? offset + executions.length
          : offset + rowsPerPage);

      setTotalCount(total);
    } catch (error) {
      console.error("Failed to load executions:", error);
      setTasks([]);
      setTotalCount(0);
    } finally {
      setLoading(false);
    }
  }, [page, rowsPerPage, orderBy, order]);

  useEffect(() => {
    loadExecutions();
  }, [loadExecutions]);

  const handleRequestSort = (property: SortableColumn) => {
    const isAsc = orderBy === property && order === "asc";
    setOrder(isAsc ? "desc" : "asc");
    setOrderBy(property);
  };

  const handleChangePage = (_: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  return (
    <Box
      sx={{
        height: "100%",
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
      }}
    >
      <Stack
        direction="row"
        sx={{
          justifyContent: "space-between",
          alignItems: "center",
          mb: 2,
          flexShrink: 0,
        }}
      >
        <Typography variant="subtitle2" sx={{ color: "text.secondary" }}>
          Active StackStorm Workflows & Event Executions
        </Typography>

        <Button
          size="small"
          variant="contained"
          startIcon={
            loading ? (
              <CircularProgress size={16} color="inherit" />
            ) : (
              <RefreshIcon />
            )
          }
          onClick={loadExecutions}
          disabled={loading}
        >
          {loading ? "Refreshing..." : "Refresh"}
        </Button>
      </Stack>

      <TableContainer
        component={Paper}
        variant="outlined"
        sx={{
          flexGrow: 1,
          minHeight: 0,
          maxHeight: "100%",
          overflow: "auto",
          ...darkScrollbarStyles,
        }}
      >
        <Table stickyHeader size="small" sx={{ minWidth: 950 }}>
          <TableHead>
            <TableRow>
              <TableCell width={50} />
              <TableCell>Execution Code</TableCell>

              <TableCell
                sortDirection={orderBy === "automation_code" ? order : false}
              >
                <TableSortLabel
                  active={orderBy === "automation_code"}
                  direction={orderBy === "automation_code" ? order : "asc"}
                  onClick={() => handleRequestSort("automation_code")}
                >
                  Job
                </TableSortLabel>
              </TableCell>

              <TableCell>Trigger Source</TableCell>
              <TableCell>Input</TableCell>

              <TableCell sortDirection={orderBy === "status" ? order : false}>
                <TableSortLabel
                  active={orderBy === "status"}
                  direction={orderBy === "status" ? order : "asc"}
                  onClick={() => handleRequestSort("status")}
                >
                  Status
                </TableSortLabel>
              </TableCell>

              <TableCell
                sortDirection={orderBy === "created_at" ? order : false}
              >
                <TableSortLabel
                  active={orderBy === "created_at"}
                  direction={orderBy === "created_at" ? order : "asc"}
                  onClick={() => handleRequestSort("created_at")}
                >
                  Start Time
                </TableSortLabel>
              </TableCell>

              <TableCell>End Time</TableCell>

              <TableCell
                sortDirection={orderBy === "duration_ms" ? order : false}
              >
                <TableSortLabel
                  active={orderBy === "duration_ms"}
                  direction={orderBy === "duration_ms" ? order : "asc"}
                  onClick={() => handleRequestSort("duration_ms")}
                >
                  Duration
                </TableSortLabel>
              </TableCell>
            </TableRow>
          </TableHead>

          <TableBody>
            {tasks.map((task) => (
              <TaskRow key={task.id} row={task} />
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <TablePagination
        rowsPerPageOptions={[10, 20, 50, 100]}
        component="div"
        count={totalCount}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
        sx={{ borderTop: "1px solid #30363d", flexShrink: 0 }}
      />
    </Box>
  );
};
