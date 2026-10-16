import React, { useEffect, useState } from "react";
import {
  Box,
  Grid,
  Paper,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Typography,
  Chip,
  Divider,
  IconButton,
  Stack,
  Tooltip,
  Breadcrumbs,
  Link,
  CircularProgress,
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
} from "@mui/material";

import {
  Refresh as RefreshIcon,
  Download as DownloadIcon,
  ContentCopy as CopyIcon,
  ArrowBack as BackIcon,
  FolderOpen as OpenFolderIcon,
  Edit as EditIcon,
  Save as SaveIcon,
  Close as CancelIcon,
} from "@mui/icons-material";

import { useFiles } from "../hooks/useFiles";
import { getFileIcon } from "./FileIcon";
import { ContainerFile, ContainerInfo } from "../models/k8sResource";
import { darkScrollbarStyles } from "../theme";
import CircularProgressWithLabel from "./CircularProgressWithLabel";
import { formatFileSize } from "../utils/format";
import { DiffReviewDialog } from "./ReviewFileChange";
import { Editor } from "@monaco-editor/react";

const getMonacoLanguage = (fileName: string): string => {
  const ext = fileName.split(".").pop()?.toLowerCase();
  switch (ext) {
    case "yaml":
    case "yml":
      return "yaml";
    case "json":
      return "json";
    case "sh":
    case "bash":
      return "shell";
    case "js":
    case "jsx":
      return "javascript";
    case "ts":
    case "tsx":
      return "typescript";
    case "py":
      return "python";
    case "xml":
      return "xml";
    case "html":
      return "html";
    case "css":
      return "css";
    case "md":
      return "markdown";
    default:
      return "plaintext";
  }
};

export const ContainerFileExplorerTab: React.FC<{
  podNamespace: string;
  podName: string;
  activeContainer?: ContainerInfo;
}> = ({ podNamespace, podName, activeContainer }) => {
  const [backupDialogOpen, setBackupDialogOpen] = useState(false);
  const [discardDialogOpen, setDiscardDialogOpen] = useState(false);
  const [diffDialogOpen, setDiffDialogOpen] = useState(false);

  const {
    files,
    fileContent,
    loadingFiles,
    loadingContent,
    saving,
    error: fileError,
    fetchFiles,
    fetchFileContent,
    saveFile,
    handleDownload,
    isDownloading,
    downloadProgress,
    currentPath,
    setCurrentPath,
    selectedFile,
    setSelectedFile,
  } = useFiles({
    namespace: podNamespace,
    pod: podName,
    container: activeContainer?.name || "",
  });

  useEffect(() => {
    if (selectedFile && !selectedFile.isDirectory) {
      fetchFileContent(selectedFile.path);
      setIsEditing(false);
    }
  }, [selectedFile]);

  // File editor state
  const [isEditing, setIsEditing] = useState<boolean>(false);
  const [editedContent, setEditedContent] = useState<string>("");

  const handleCancelEdit = () => {
    const hasUnsavedChanges = editedContent !== (fileContent || "");

    if (hasUnsavedChanges) {
      setDiscardDialogOpen(true);
    } else {
      setEditedContent(fileContent || "");
      setIsEditing(false);
    }
  };

  const confirmDiscard = () => {
    setEditedContent(fileContent || "");
    setIsEditing(false);
    setDiscardDialogOpen(false);
  };

  const handleEnterEditing = () => {
    if (selectedFile) {
      setBackupDialogOpen(true);
    } else {
      setIsEditing(true);
    }
  };

  const confirmBackup = async () => {
    if (selectedFile) {
      await handleDownload(selectedFile);
    }
    setIsEditing(true);
    setBackupDialogOpen(false);
  };

  const skipBackup = () => {
    setIsEditing(true);
    setBackupDialogOpen(false);
  };

  const handleOpenDiff = () => {
    if (!selectedFile) return;
    setDiffDialogOpen(true);
  };

  const handleConfirmSave = async () => {
    if (!selectedFile) return;

    const success = await saveFile(selectedFile.path, editedContent);

    if (success) {
      setDiffDialogOpen(false);
      setIsEditing(false);
      fetchFileContent(selectedFile.path);
    }
  };

  const hasUnsavedChanges = isEditing && editedContent !== fileContent;

  useEffect(() => {
    setEditedContent(fileContent || "");
  }, [fileContent]);

  const handleOpenFolder = (folderPath: string) => {
    const formattedPath = folderPath.endsWith("/")
      ? folderPath
      : `${folderPath}/`;
    setCurrentPath(formattedPath);
    setSelectedFile(null);
  };

  function fileEditor(
    loadingContent: boolean,
    isEditing: boolean,
    selectedFile: ContainerFile,
  ): React.ReactNode {
    if (loadingContent) {
      return (
        <Stack
          direction="row"
          spacing={1}
          sx={{
            alignItems: "center",
            justifyContent: "center",
            height: "100%",
            color: "text.secondary",
          }}
        >
          <CircularProgress size={20} color="inherit" />
          <Typography variant="body2">Loading file stream...</Typography>
        </Stack>
      );
    }

    const language = getMonacoLanguage(selectedFile.name);

    if (isEditing) {
      return (
        <Editor
          height="100%"
          width="100%"
          language={language}
          theme="vs-dark"
          value={editedContent}
          onChange={(value) => setEditedContent(value || "")}
          options={{
            fontSize: 13,
            fontFamily:
              "'Fira Code', 'Cascadia Code', 'Courier New', monospace",
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            wordWrap: "on",
            padding: { top: 10, bottom: 10 },
            lineNumbersMinChars: 4,
            cursorBlinking: "smooth",
            smoothScrolling: true,
            readOnly: false,
          }}
        />
      );
    }

    return (
      <Editor
        height="100%"
        width="100%"
        language={language}
        theme="vs-dark"
        value={fileContent || selectedFile.content || "// File is empty"}
        options={{
          readOnly: true,
          domReadOnly: true,
          fontSize: 13,
          fontFamily: "'Fira Code', 'Cascadia Code', 'Courier New', monospace",
          minimap: { enabled: false },
          scrollBeyondLastLine: false,
          automaticLayout: true,
          wordWrap: "on",
          padding: { top: 10, bottom: 10 },
          lineNumbersMinChars: 4,
          smoothScrolling: true,
        }}
      />
    );
  }

  return (
    <Box
      sx={{
        height: "100%",
        width: "100%",
        display: "flex",
        flexDirection: "column",
        minHeight: 0,
        overflow: "hidden",
      }}
    >
      <Grid
        container
        spacing={2}
        sx={{ flexGrow: 1, height: "100%", minHeight: 0, overflow: "hidden" }}
      >
        {fileError && (
          <Grid size={{ xs: 12 }} sx={{ flexShrink: 0 }}>
            <Alert severity="error" sx={{ py: 0 }}>
              {fileError}
            </Alert>
          </Grid>
        )}

        <Grid
          size={{ xs: 12, md: 5 }}
          sx={{
            height: "100%",
            display: "flex",
            flexDirection: "column",
            minHeight: 0,
            overflow: "hidden",
          }}
        >
          <Paper
            variant="outlined"
            sx={{ p: 1, mb: 1, bgcolor: "action.hover", flexShrink: 0 }}
          >
            <Stack direction="row" spacing={1} sx={{ alignItems: "center" }}>
              {currentPath !== "/" && (
                <IconButton
                  size="small"
                  onClick={() => {
                    const segments = currentPath.split("/").filter(Boolean);
                    segments.pop();
                    const parentPath =
                      segments.length > 0 ? `/${segments.join("/")}/` : "/";
                    setCurrentPath(parentPath);
                    setSelectedFile(null);
                  }}
                >
                  <BackIcon fontSize="small" />
                </IconButton>
              )}
              <OpenFolderIcon fontSize="small" color="action" />
              <Breadcrumbs
                maxItems={3}
                aria-label="breadcrumb"
                sx={{ fontSize: "0.8rem" }}
              >
                <Link
                  underline="hover"
                  color="inherit"
                  sx={{ cursor: "pointer" }}
                  onClick={() => {
                    setCurrentPath("/");
                    setSelectedFile(null);
                  }}
                >
                  root
                </Link>
                {currentPath
                  .split("/")
                  .filter(Boolean)
                  .map((folder, idx, arr) => {
                    const subPath = "/" + arr.slice(0, idx + 1).join("/") + "/";
                    return (
                      <Link
                        key={subPath}
                        underline="hover"
                        color={
                          idx === arr.length - 1 ? "text.primary" : "inherit"
                        }
                        sx={{ cursor: "pointer" }}
                        onClick={() => {
                          setCurrentPath(subPath);
                          setSelectedFile(null);
                        }}
                      >
                        {folder}
                      </Link>
                    );
                  })}
              </Breadcrumbs>
              <IconButton size="small" onClick={fetchFiles} sx={{ ml: "auto" }}>
                <RefreshIcon fontSize="small" />
              </IconButton>
            </Stack>
          </Paper>

          <Paper
            variant="outlined"
            sx={{
              flexGrow: 1,
              minHeight: 0,
              overflow: "auto",
              ...darkScrollbarStyles,
            }}
          >
            {loadingFiles ? (
              <Stack
                sx={{ alignItems: "center", justifyContent: "center", py: 4 }}
              >
                <CircularProgress size={24} />
                <Typography variant="caption" color="text.secondary">
                  Fetching folder content...
                </Typography>
              </Stack>
            ) : (
              <List dense>
                {files && files.length > 0 ? (
                  files.map((file, idx) => (
                    <ListItemButton
                      key={file.path || idx}
                      selected={selectedFile?.path === file.path}
                      onClick={() => {
                        if (file.isDirectory) {
                          handleOpenFolder(file.path);
                        } else {
                          setSelectedFile(file);
                        }
                      }}
                    >
                      <ListItemIcon sx={{ minWidth: 36 }}>
                        {getFileIcon(file)}
                      </ListItemIcon>
                      <ListItemText
                        primary={file.name}
                        secondary={
                          file.isDirectory
                            ? "Directory"
                            : `${formatFileSize(file.size || 0)} ${
                                file.permissions ? `• ${file.permissions}` : ""
                              }`
                        }
                        slotProps={{
                          primary: {
                            variant: "body2",
                            sx: { fontWeight: file.isDirectory ? 600 : 400 },
                          },
                        }}
                      />
                    </ListItemButton>
                  ))
                ) : (
                  <Typography
                    variant="caption"
                    sx={{ p: 2, display: "block", color: "text.secondary" }}
                  >
                    No items inside <code>{currentPath}</code>
                  </Typography>
                )}
              </List>
            )}
          </Paper>
        </Grid>

        {/* File Details & Editor Panel */}
        <Grid
          size={{ xs: 12, md: 7 }}
          sx={{
            height: "100%",
            display: "flex",
            flexDirection: "column",
            minHeight: 0,
            overflow: "hidden",
          }}
        >
          {selectedFile ? (
            <Paper
              variant="outlined"
              sx={{
                p: 2,
                height: "100%",
                display: "flex",
                flexDirection: "column",
                minHeight: 0,
                overflow: "hidden",
              }}
            >
              <Box sx={{ flexShrink: 0 }}>
                <Stack
                  direction="row"
                  sx={{
                    justifyContent: "space-between",
                    alignItems: "flex-start",
                    mb: 1,
                  }}
                >
                  <Box>
                    <Stack
                      direction="row"
                      spacing={1}
                      sx={{ alignItems: "center", mb: 0.5 }}
                    >
                      <Typography
                        variant="subtitle1"
                        sx={{ fontWeight: 600, fontFamily: "monospace" }}
                      >
                        {selectedFile.name}
                      </Typography>
                      {hasUnsavedChanges && (
                        <Chip
                          label="Unsaved Changes"
                          size="small"
                          color="warning"
                          variant="outlined"
                        />
                      )}
                    </Stack>
                    <Typography variant="caption" color="text.secondary">
                      Path: <code>{selectedFile.path}</code>
                    </Typography>
                  </Box>

                  <Stack
                    direction="row"
                    spacing={1}
                    sx={{ alignItems: "center" }}
                  >
                    {isEditing ? (
                      <>
                        <Button
                          size="small"
                          variant="outlined"
                          color="inherit"
                          startIcon={<CancelIcon />}
                          onClick={handleCancelEdit}
                          disabled={saving}
                        >
                          Cancel
                        </Button>

                        <Button
                          size="small"
                          variant="contained"
                          color="primary"
                          startIcon={<SaveIcon />}
                          onClick={handleOpenDiff}
                          disabled={saving || !hasUnsavedChanges}
                        >
                          Review & Save
                        </Button>
                      </>
                    ) : (
                      <>
                        <Button
                          size="small"
                          variant="outlined"
                          startIcon={<EditIcon />}
                          onClick={handleEnterEditing}
                          disabled={loadingContent || !selectedFile.editable}
                        >
                          Edit
                        </Button>

                        {isDownloading ? (
                          <CircularProgressWithLabel
                            variant="determinate"
                            value={downloadProgress}
                          />
                        ) : (
                          <Tooltip title="Download File">
                            <IconButton
                              size="small"
                              color="primary"
                              onClick={() => handleDownload(selectedFile)}
                              disabled={loadingContent}
                            >
                              <DownloadIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        )}

                        <Tooltip title="Copy File Path">
                          <IconButton
                            size="small"
                            onClick={() =>
                              navigator.clipboard.writeText(selectedFile.path)
                            }
                          >
                            <CopyIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </>
                    )}
                  </Stack>
                </Stack>

                <Stack direction="row" spacing={1} sx={{ mb: 1.5 }}>
                  <Chip
                    label={`Size: ${formatFileSize(selectedFile.size || 0)}`}
                    size="small"
                    variant="filled"
                  />
                  {selectedFile.permissions && (
                    <Chip
                      label={`Mode: ${selectedFile.permissions}`}
                      size="small"
                      variant="filled"
                    />
                  )}
                  {selectedFile.updatedAt && (
                    <Chip
                      label={`Modified: ${selectedFile.updatedAt}`}
                      size="small"
                      variant="filled"
                    />
                  )}
                  <Chip
                    label={selectedFile.editable ? "Editable" : "Read-Only"}
                    color={selectedFile.editable ? "success" : "error"}
                    size="small"
                    variant="filled"
                  />
                </Stack>

                <Divider sx={{ mb: 1.5 }} />
              </Box>

              {/* Clean Monaco Editor Wrapper Box */}
              <Box
                sx={{
                  flexGrow: 1,
                  minHeight: 0,
                  height: "100%",
                  width: "100%",
                  borderRadius: 1,
                  overflow: "hidden",
                  border: "1px solid",
                  borderColor: "divider",
                  bgcolor: "#1e1e1e",
                }}
              >
                {fileEditor(loadingContent, isEditing, selectedFile)}
              </Box>
            </Paper>
          ) : (
            <Paper
              variant="outlined"
              sx={{
                p: 2,
                height: "100%",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                color: "text.secondary",
              }}
            >
              <Typography variant="caption">
                Select a file from the list to preview details & download
              </Typography>
            </Paper>
          )}
        </Grid>
      </Grid>

      <DiffReviewDialog
        open={diffDialogOpen}
        fileName={selectedFile?.name || ""}
        originalContent={fileContent || ""}
        modifiedContent={editedContent}
        saving={saving}
        onClose={() => setDiffDialogOpen(false)}
        onConfirmSave={handleConfirmSave}
      />

      {/* Backup Dialog */}
      <Dialog
        open={backupDialogOpen}
        onClose={() => setBackupDialogOpen(false)}
      >
        <DialogTitle>Backup File?</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Would you like to download a backup of{" "}
            <strong>{selectedFile?.name}</strong> before editing?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setBackupDialogOpen(false)} color="inherit">
            Cancel
          </Button>
          <Button onClick={skipBackup} color="inherit">
            Skip
          </Button>
          <Button
            onClick={confirmBackup}
            variant="contained"
            color="primary"
            autoFocus
          >
            Download Backup
          </Button>
        </DialogActions>
      </Dialog>

      {/* Discard Dialog */}
      <Dialog
        open={discardDialogOpen}
        onClose={() => setDiscardDialogOpen(false)}
      >
        <DialogTitle>Discard Unsaved Changes?</DialogTitle>
        <DialogContent>
          <DialogContentText>
            You have unsaved changes in this file. Are you sure you want to
            discard them?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDiscardDialogOpen(false)} color="inherit">
            Cancel
          </Button>
          <Button
            onClick={confirmDiscard}
            variant="contained"
            color="error"
            autoFocus
          >
            Discard Changes
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
