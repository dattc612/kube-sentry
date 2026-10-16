import React, { useState } from "react";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import CircularProgress from "@mui/material/CircularProgress";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import ErrorIcon from "@mui/icons-material/Error";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import Button from "@mui/material/Button";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

import { useJobs } from "../hooks/useJobs";
import { JobItem } from "../models/jobs";
import { ContainerInfo } from "../models/k8sResource";

export const RunJobListItem: React.FC<{
  podNamespace: string;
  podName: string;
  activeContainer?: ContainerInfo;
}> = ({ podNamespace, podName, activeContainer }) => {
  const { jobs, status, loadingJobs, running, error, output, handleRunJob } =
    useJobs({
      namespace: podNamespace,
      pod: podName,
      container: activeContainer?.name || "",
    });

  const [openDialog, setOpenDialog] = useState(false);
  const [selectedJob, setSelectedJob] = useState<JobItem | null>(null);
  const [activeJobId, setActiveJobId] = useState<number | null>(null);

  const handleOnClickRunBtn = (job: JobItem) => {
    setSelectedJob(job);
    setOpenDialog(true);
  };

  const handleConfirmRun = async () => {
    if (selectedJob) {
      const jobIdToRun = selectedJob.ID;
      setActiveJobId(jobIdToRun);
      setOpenDialog(false);
      setSelectedJob(null);
      await handleRunJob(jobIdToRun);
    } else {
      setOpenDialog(false);
    }
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setSelectedJob(null);
  };

  return (
    <>
      <List>
        {jobs?.map((job: JobItem) => {
          const isActive = job.ID === activeJobId;
          const isRunning = isActive && running;
          const isSuccess = isActive && status === "success";
          const isError = isActive && status === "error";

          return (
            <ListItem
              disablePadding
              key={job.ID}
              sx={{ flexDirection: "column", alignItems: "stretch" }}
            >
              <ListItemButton
                onClick={() => handleOnClickRunBtn(job)}
                disabled={loadingJobs || running}
              >
                <ListItemIcon>
                  {isRunning ? (
                    <CircularProgress size={24} />
                  ) : isSuccess ? (
                    <CheckCircleIcon color="success" />
                  ) : isError || (isActive && error) ? (
                    <ErrorIcon color="error" />
                  ) : (
                    <PlayArrowIcon />
                  )}
                </ListItemIcon>
                <ListItemText
                  primary={`${job.JobName}`}
                  secondary={
                    isRunning
                      ? "Processing..."
                      : isSuccess
                        ? "Completed"
                        : isError
                          ? "Failed"
                          : job.Description
                  }
                />
              </ListItemButton>

              {isActive && (
                <Box sx={{ pl: 9, pr: 2, pb: 2, pt: 0.5 }}>
                  <Box
                    sx={{
                      bgcolor: "#0d1117",
                      borderRadius: 2,
                      border: "1px solid #30363d",
                      overflow: "hidden",
                      boxShadow: "0 4px 12px rgba(0,0,0,0.5)",
                    }}
                  >
                    <Box
                      sx={{
                        bgcolor: "#161b22",
                        px: 1.5,
                        py: 0.75,
                        display: "flex",
                        alignItems: "center",
                        borderBottom: "1px solid #30363d",
                      }}
                    >
                      <Box sx={{ display: "flex", gap: 0.75, mr: 1.5 }}>
                        <Box
                          sx={{
                            width: 10,
                            height: 10,
                            borderRadius: "50%",
                            bgcolor: "#ff5f56",
                          }}
                        />
                        <Box
                          sx={{
                            width: 10,
                            height: 10,
                            borderRadius: "50%",
                            bgcolor: "#ffbd2e",
                          }}
                        />
                        <Box
                          sx={{
                            width: 10,
                            height: 10,
                            borderRadius: "50%",
                            bgcolor: "#27c93f",
                          }}
                        />
                      </Box>
                      <Typography
                        variant="caption"
                        sx={{
                          color: "#8b949e",
                          fontFamily: "'Fira Code', 'Roboto Mono', monospace",
                          fontSize: "0.75rem",
                          fontWeight: 500,
                        }}
                      >
                        {job.JobName} — Output
                      </Typography>
                    </Box>

                    {/* Terminal Window Body */}
                    <Box
                      sx={{
                        p: 1.5,
                        maxHeight: 280,
                        overflowY: "auto",
                        fontFamily:
                          "'Fira Code', 'Consolas', 'Courier New', monospace",
                        fontSize: "0.8125rem",
                        lineHeight: 1.6,
                        "&::-webkit-scrollbar": {
                          width: "6px",
                        },
                        "&::-webkit-scrollbar-thumb": {
                          backgroundColor: "#30363d",
                          borderRadius: "3px",
                        },
                      }}
                    >
                      {error && (
                        <Box
                          component="pre"
                          sx={{
                            color: "#f85149",
                            margin: 0,
                            whiteSpace: "pre-wrap",
                            wordBreak: "break-word",
                          }}
                        >
                          <span
                            style={{ color: "#da3633", fontWeight: "bold" }}
                          >
                            [ERROR]:{" "}
                          </span>
                          {error}
                        </Box>
                      )}
                      {output && (
                        <Box
                          component="pre"
                          sx={{
                            color: "#3fb950",
                            margin: 0,
                            whiteSpace: "pre-wrap",
                            wordBreak: "break-word",
                          }}
                        >
                          {output}
                        </Box>
                      )}
                    </Box>
                  </Box>
                </Box>
              )}
            </ListItem>
          );
        })}
      </List>

      <Dialog
        open={openDialog}
        onClose={handleCloseDialog}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogTitle id="alert-dialog-title">Confirm Job Run</DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            Do you want to run job <strong>{selectedJob?.JobName}</strong> in{" "}
            <strong>{activeContainer?.name}</strong> of{" "}
            <strong>{podName}</strong>?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog} color="inherit">
            Cancel
          </Button>
          <Button
            onClick={handleConfirmRun}
            variant="contained"
            color="primary"
            autoFocus
          >
            Run Job
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};
