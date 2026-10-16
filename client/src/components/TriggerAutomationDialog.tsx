import React, { useEffect, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
} from "@mui/material";
import Editor from "@monaco-editor/react";
import { Automation } from "../models/automation";

interface TriggerAutomationDialogProps {
  open: boolean;
  automation: Automation | null;
  loading?: boolean;
  onClose: () => void;
  onTrigger: (input: Record<string, unknown>) => Promise<void>;
}

export const TriggerAutomationDialog: React.FC<
  TriggerAutomationDialogProps
> = ({ open, automation, loading = false, onClose, onTrigger }) => {
  const [value, setValue] = useState("{}");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (automation) {
      setValue(JSON.stringify(automation.sample_input ?? {}, null, 2));
      setError(null);
    }
  }, [automation]);

  const handleTrigger = async () => {
    try {
      setError(null);

      const input = JSON.parse(value);

      if (typeof input !== "object" || input === null || Array.isArray(input)) {
        setError("Input must be a JSON object.");
        return;
      }

      await onTrigger(input as Record<string, unknown>);
    } catch (err) {
      if (err instanceof SyntaxError) {
        setError("Invalid JSON. Please fix the JSON before triggering.");
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to trigger automation.");
      }
    }
  };

  return (
    <Dialog
      open={open}
      onClose={loading ? undefined : onClose}
      fullWidth
      maxWidth="lg"
    >
      <DialogTitle>Trigger: {automation?.name}</DialogTitle>

      <DialogContent>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <Box
          sx={{
            height: 500,
            border: "1px solid",
            borderColor: "divider",
            borderRadius: 1,
            overflow: "hidden",
          }}
        >
          <Editor
            height="100%"
            language="json"
            theme="vs-dark"
            value={value}
            onChange={(newValue) => {
              setValue(newValue ?? "");
              setError(null);
            }}
            options={{
              minimap: {
                enabled: false,
              },
              formatOnPaste: true,
              formatOnType: true,
              automaticLayout: true,
              tabSize: 2,
              scrollBeyondLastLine: false,
            }}
          />
        </Box>
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose} disabled={loading}>
          Cancel
        </Button>

        <Button
          variant="contained"
          onClick={handleTrigger}
          disabled={loading || !automation}
          startIcon={loading ? undefined : undefined}
        >
          {loading ? "Queuing..." : "Trigger"}
        </Button>
      </DialogActions>
    </Dialog>
  );
};
