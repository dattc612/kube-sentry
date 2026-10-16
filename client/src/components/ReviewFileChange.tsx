import React from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Stack,
  Typography,
  Chip,
  CircularProgress,
} from "@mui/material";
import { Save as SaveIcon, Close as CloseIcon } from "@mui/icons-material";
import { DiffEditor } from "@monaco-editor/react";

interface DiffReviewDialogProps {
  open: boolean;
  fileName: string;
  originalContent: string;
  modifiedContent: string;
  saving?: boolean;
  onClose: () => void;
  onConfirmSave: () => void;
}

const getLanguage = (fileName: string): string => {
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
    case "ts":
    case "tsx":
      return "typescript";
    case "py":
      return "python";
    case "xml":
      return "xml";
    default:
      return "plaintext";
  }
};

export const DiffReviewDialog: React.FC<DiffReviewDialogProps> = ({
  open,
  fileName,
  originalContent,
  modifiedContent,
  saving = false,
  onClose,
  onConfirmSave,
}) => {
  const language = getLanguage(fileName);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="lg"
      fullWidth
      slotProps={{
        transition: {
          unmountOnExit: true,
        },
      }}
    >
      <DialogTitle
        sx={{ borderBottom: "1px solid", borderColor: "divider", py: 1.5 }}
      >
        <Stack
          direction="row"
          sx={{ justifyContent: "space-between", alignItems: "center" }}
        >
          <Stack direction="row" spacing={1} sx={{ alignItems: "center" }}>
            <Typography variant="h6" sx={{ fontSize: "1rem", fontWeight: 600 }}>
              Review Changes: {fileName}
            </Typography>
            <Chip
              label={language.toUpperCase()}
              size="small"
              variant="outlined"
              color="primary"
            />
          </Stack>
          <Typography variant="caption" color="text.secondary">
            Original (Left) vs. Unsaved Changes (Right)
          </Typography>
        </Stack>
      </DialogTitle>

      <DialogContent sx={{ p: 0, height: "500px", bgcolor: "#1e1e1e" }}>
        {open && (
          <DiffEditor
            height="100%"
            language={language}
            original={originalContent}
            modified={modifiedContent}
            theme="vs-dark"
            keepCurrentModifiedModel={true}
            options={{
              readOnly: true,
              renderSideBySide: true,
              minimap: { enabled: false },
              scrollBeyondLastLine: false,
              fontSize: 13,
              automaticLayout: true,
            }}
          />
        )}
      </DialogContent>

      <DialogActions
        sx={{ p: 1.5, borderTop: "1px solid", borderColor: "divider" }}
      >
        <Button
          onClick={onClose}
          color="inherit"
          startIcon={<CloseIcon />}
          disabled={saving}
        >
          Cancel
        </Button>
        <Button
          onClick={onConfirmSave}
          variant="contained"
          color="primary"
          startIcon={
            saving ? (
              <CircularProgress size={16} color="inherit" />
            ) : (
              <SaveIcon />
            )
          }
          disabled={saving}
        >
          Confirm & Save
        </Button>
      </DialogActions>
    </Dialog>
  );
};
