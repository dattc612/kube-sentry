import React, { useState, useEffect, useRef } from "react";
import {
  ThemeProvider,
  createTheme,
  CssBaseline,
  Box,
  AppBar,
  Toolbar,
  Typography,
  Button,
  CircularProgress,
} from "@mui/material";
import LogoutIcon from "@mui/icons-material/Logout";
import Dashboard from "./pages/Dashboard";
import { Login } from "./pages/Login";
import keycloak from "./auth/keycloak";

const theme = createTheme({
  palette: {
    mode: "dark",
    background: {
      default: "#0d1117",
      paper: "#161b22",
    },
    primary: {
      main: "#58a6ff",
    },
  },
  components: {
    MuiAppBar: {
      styleOverrides: {
        root: ({ theme }) => ({
          backgroundColor: theme.palette.background.paper,
          color: theme.palette.text.primary,
          borderBottom: `1px solid ${theme.palette.divider}`,
          boxShadow: "none",
        }),
      },
    },
    MuiToolbar: {
      styleOverrides: {
        dense: {
          minHeight: 48,
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: "none",
        },
      },
    },
  },
});

type AuthProvider = "local" | "keycloak" | null;

export default function App(): React.JSX.Element {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [authProvider, setAuthProvider] = useState<AuthProvider>(null);
  const [loading, setLoading] = useState<boolean>(true);

  const isInitialized = useRef<boolean>(false);

  useEffect(() => {
    if (isInitialized.current) return;
    isInitialized.current = true;

    const savedProvider = localStorage.getItem("authProvider") as AuthProvider;
    const localToken = localStorage.getItem("authToken");

    if (savedProvider === "local" && localToken) {
      setIsAuthenticated(true);
      setAuthProvider("local");
      setLoading(false);
      return;
    }

    keycloak
      .init({
        onLoad: "check-sso",
        pkceMethod: "S256",
        checkLoginIframe: false,
      })
      .then((authenticated: boolean) => {
        if (authenticated && keycloak.token) {
          localStorage.setItem("authToken", keycloak.token);
          localStorage.setItem("authProvider", "keycloak");

          setIsAuthenticated(true);
          setAuthProvider("keycloak");

          keycloak.onTokenExpired = () => {
            keycloak
              .updateToken(70)
              .then((refreshed) => {
                if (refreshed && keycloak.token) {
                  localStorage.setItem("authToken", keycloak.token);
                }
              })
              .catch(() => handleLogout());
          };
        } else {
          localStorage.removeItem("authToken");
          localStorage.removeItem("authProvider");

          setIsAuthenticated(false);
          setAuthProvider(null);
        }
      })
      .catch((err: unknown) => {
        console.error("Keycloak initialization error:", err);
        setIsAuthenticated(false);
        setAuthProvider(null);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  const handleLoginSuccess = (): void => {
    localStorage.setItem("authProvider", "local");

    setIsAuthenticated(true);
    setAuthProvider("local");
  };

  const handleKeycloakLogin = (): void => {
    keycloak.login({
      redirectUri: window.location.origin,
    });
  };

  const handleLogout = (): void => {
    localStorage.removeItem("authToken");
    localStorage.removeItem("authProvider");

    setIsAuthenticated(false);
    setAuthProvider(null);
  };

  if (loading) {
    return (
      <ThemeProvider theme={theme}>
        <CssBaseline />

        <Box
          sx={{
            display: "flex",
            height: "100vh",
            justifyContent: "center",
            alignItems: "center",
            bgcolor: "background.default",
          }}
        >
          <CircularProgress color="primary" />
        </Box>
      </ThemeProvider>
    );
  }

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />

      {!isAuthenticated ? (
        <Login
          onLoginSuccess={handleLoginSuccess}
          onKeycloakLogin={handleKeycloakLogin}
        />
      ) : (
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
            height: "100vh",
            maxHeight: "100vh",
            width: "100vw",
            overflow: "hidden",
            bgcolor: "background.default",
          }}
        >
          <AppBar position="static" color="default" elevation={0}>
            <Toolbar
              variant="dense"
              sx={{
                justifyContent: "space-between",
                px: 2,
              }}
            >
              <Typography
                variant="h6"
                component="div"
                sx={{
                  fontSize: "1rem",
                  fontWeight: 600,
                  letterSpacing: 0.2,
                }}
              >
                Go Through the Storm
              </Typography>

              <Button
                color="inherit"
                size="small"
                onClick={handleLogout}
                startIcon={<LogoutIcon />}
                sx={{
                  color: "text.secondary",
                  "&:hover": {
                    color: "text.primary",
                    bgcolor: "action.hover",
                  },
                }}
              >
                Logout
              </Button>
            </Toolbar>
          </AppBar>

          <Box
            sx={{
              flexGrow: 1,
              minHeight: 0,
              display: "flex",
              flexDirection: "column",
              overflow: "hidden",
            }}
          >
            <Dashboard />
          </Box>
        </Box>
      )}
    </ThemeProvider>
  );
}
