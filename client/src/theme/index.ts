import { createTheme } from "@mui/material/styles";

export const theme = createTheme({
  palette: {
    mode: "dark",

    background: {
      default: "#080B10",
      paper: "#0F141C",
    },

    primary: {
      main: "#4DA3FF",
      light: "#79BBFF",
      dark: "#2477C9",
      contrastText: "#FFFFFF",
    },

    secondary: {
      main: "#8B9BB4",
      light: "#AEBBCB",
      dark: "#59687D",
      contrastText: "#FFFFFF",
    },

    success: {
      main: "#3DDC97",
    },

    warning: {
      main: "#FFB547",
    },

    error: {
      main: "#FF5C67",
    },

    info: {
      main: "#4DA3FF",
    },

    text: {
      primary: "#E7ECF3",
      secondary: "#8D9AAC",
      disabled: "#566273",
    },

    divider: "#202936",
  },

  typography: {
    fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
    fontSize: 13,

    h1: {
      fontWeight: 600,
      letterSpacing: "-0.02em",
    },
    h2: {
      fontWeight: 600,
      letterSpacing: "-0.02em",
    },
    h3: {
      fontWeight: 600,
      letterSpacing: "-0.015em",
    },
    h4: {
      fontWeight: 600,
    },
    h5: {
      fontWeight: 600,
    },
    h6: {
      fontWeight: 600,
    },

    button: {
      fontWeight: 600,
      textTransform: "none",
    },
  },

  shape: {
    borderRadius: 6,
  },

  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          backgroundColor: "#080B10",
          color: "#E7ECF3",
        },
        "::selection": {
          backgroundColor: "#4DA3FF",
          color: "#FFFFFF",
        },
      },
    },

    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: "none",
          backgroundColor: "#0F141C",
          border: "1px solid #202936",
          boxShadow: "0 4px 20px rgba(0, 0, 0, 0.25)",
        },
      },
    },

    MuiCard: {
      styleOverrides: {
        root: {
          backgroundColor: "#0F141C",
          backgroundImage: "none",
          border: "1px solid #202936",
          boxShadow: "0 4px 20px rgba(0, 0, 0, 0.25)",
          transition: "border-color 150ms ease, box-shadow 150ms ease",

          "&:hover": {
            borderColor: "#303D4E",
            boxShadow: "0 6px 24px rgba(0, 0, 0, 0.35)",
          },
        },
      },
    },

    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 5,
          textTransform: "none",
          fontWeight: 600,
        },

        contained: {
          boxShadow: "none",

          "&:hover": {
            boxShadow: "0 0 16px rgba(77, 163, 255, 0.18)",
          },
        },

        outlined: {
          borderColor: "#303B4A",

          "&:hover": {
            borderColor: "#4DA3FF",
            backgroundColor: "rgba(77, 163, 255, 0.06)",
          },
        },
      },

      variants: [
        {
          props: { variant: "contained", color: "primary" },
          style: {
            backgroundColor: "#4DA3FF",

            "&:hover": {
              backgroundColor: "#66AEFF",
            },
          },
        },
      ],
    },

    MuiTextField: {
      defaultProps: {
        variant: "outlined",
      },
    },

    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          backgroundColor: "#0A0F16",
          borderRadius: 5,

          "& .MuiOutlinedInput-notchedOutline": {
            borderColor: "#293442",
          },

          "&:hover .MuiOutlinedInput-notchedOutline": {
            borderColor: "#3B4A5C",
          },

          "&.Mui-focused .MuiOutlinedInput-notchedOutline": {
            borderColor: "#4DA3FF",
            borderWidth: 1,
          },
        },
      },
    },

    MuiInputLabel: {
      styleOverrides: {
        root: {
          color: "#7F8C9E",

          "&.Mui-focused": {
            color: "#4DA3FF",
          },
        },
      },
    },

    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: "#0A0E14",
          backgroundImage: "none",
          borderBottom: "1px solid #202936",
          boxShadow: "none",
        },
      },
    },

    MuiDrawer: {
      styleOverrides: {
        paper: {
          backgroundColor: "#0A0E14",
          backgroundImage: "none",
          borderRight: "1px solid #202936",
        },
      },
    },

    MuiDivider: {
      styleOverrides: {
        root: {
          borderColor: "#202936",
        },
      },
    },

    MuiChip: {
      styleOverrides: {
        root: ({ ownerState }) => ({
          borderRadius: 4,
          border: "1px solid #293442",

          ...(ownerState.color === "default" && {
            backgroundColor: "#171E28",
          }),

          ...(ownerState.color === "success" && {
            backgroundColor: "rgba(61, 220, 151, 0.15)",
            color: "#3DDC97",
            borderColor: "rgba(61, 220, 151, 0.3)",
          }),

          ...(ownerState.color === "error" && {
            backgroundColor: "rgba(255, 92, 103, 0.15)",
            color: "#FF5C67",
            borderColor: "rgba(255, 92, 103, 0.3)",
          }),
        }),
      },
    },

    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          backgroundColor: "#161D27",
          border: "1px solid #303B4A",
          color: "#E7ECF3",
          fontSize: 12,
        },
      },
    },

    MuiTableCell: {
      styleOverrides: {
        root: {
          borderColor: "#202936",
        },

        head: {
          backgroundColor: "#0A0F16",
          color: "#9AA8BA",
          fontWeight: 600,
        },
      },
    },
  },
});

export const darkScrollbarStyles = {
  "&::-webkit-scrollbar": {
    width: "6px",
    height: "6px",
  },

  "&::-webkit-scrollbar-thumb": {
    backgroundColor: "#303B4A",
    borderRadius: "3px",

    "&:hover": {
      backgroundColor: "#405064",
    },
  },

  "&::-webkit-scrollbar-track": {
    backgroundColor: "#080B10",
  },

  scrollbarColor: "#303B4A #080B10",
  scrollbarWidth: "thin",
};
