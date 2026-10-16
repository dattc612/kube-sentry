import React, { useState } from "react";
import {
  Box,
  Button,
  TextField,
  Paper,
  Typography,
  Alert,
  Divider,
  CircularProgress,
} from "@mui/material";
import { authApi } from "../api/authApi";

interface LoginProps {
  onLoginSuccess: () => void;
  onKeycloakLogin: () => void;
}

export const Login: React.FC<LoginProps> = ({
  onLoginSuccess,
  onKeycloakLogin,
}) => {
  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

  const handleLegacyLogin = async (
    e: React.FormEvent<HTMLFormElement>,
  ): Promise<void> => {
    e.preventDefault();
    setError("");

    if (!username.trim() || !password) {
      setError("Please enter both username and password.");
      return;
    }

    setIsSubmitting(true);

    try {
      const res = await authApi.login(username, password);
      localStorage.setItem("authToken", res.token);
      onLoginSuccess();
    } catch (err: any) {
      setError(
        err?.response?.data?.error ||
          err?.message ||
          "Failed to authenticate. Please check your credentials.",
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Box
      sx={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        minHeight: "100vh",
        bgcolor: "#0d1117",
      }}
    >
      <Paper
        variant="outlined"
        sx={{
          p: 4,
          width: 320,
          display: "flex",
          flexDirection: "column",
          gap: 2,
        }}
      >
        <Typography variant="h6" sx={{ fontWeight: 700, textAlign: "center" }}>
          Sign In
        </Typography>

        {error && <Alert severity="error">{error}</Alert>}

        <Button
          variant="contained"
          color="primary"
          size="large"
          fullWidth
          onClick={onKeycloakLogin}
          sx={{ py: 1.2, fontWeight: "bold" }}
        >
          Sign In with Keycloak
        </Button>

        <Divider sx={{ my: 1 }}>OR</Divider>

        <Box
          component="form"
          onSubmit={handleLegacyLogin}
          sx={{ display: "flex", flexDirection: "column", gap: 2 }}
        >
          <TextField
            label="Username"
            size="small"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            fullWidth
            required
            disabled={isSubmitting}
            slotProps={{
              htmlInput: { autoComplete: "username" },
            }}
          />
          <TextField
            label="Password"
            type="password"
            size="small"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            fullWidth
            required
            disabled={isSubmitting}
            slotProps={{
              htmlInput: { autoComplete: "current-password" },
            }}
          />
          <Button
            type="submit"
            variant="outlined"
            fullWidth
            disabled={isSubmitting}
            startIcon={
              isSubmitting ? (
                <CircularProgress size={16} color="inherit" />
              ) : null
            }
          >
            {isSubmitting ? "Authenticating..." : "Login with Credentials"}
          </Button>
        </Box>
      </Paper>
    </Box>
  );
};
