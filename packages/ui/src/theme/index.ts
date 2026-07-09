import { createTheme } from "@mui/material/styles";

export const linoraColors = {
  primary: "#0F8B6D",
  primaryHover: "#0B6F58",
  primarySoft: "#E3F4EF",
  charcoal: "#242826",
  ivory: "#F8F6F0",
  white: "#FFFFFF",
  border: "#E5E1D8",
  gold: "#D6B35A",
  mutedText: "#6B7280",
};

export const linoraTheme = createTheme({
  palette: {
    primary: {
      main: linoraColors.primary,
      dark: linoraColors.primaryHover,
      light: linoraColors.primarySoft,
      contrastText: linoraColors.white,
    },
    secondary: {
      main: linoraColors.gold,
      contrastText: linoraColors.charcoal,
    },
    background: {
      default: linoraColors.ivory,
      paper: linoraColors.white,
    },
    text: {
      primary: linoraColors.charcoal,
      secondary: linoraColors.mutedText,
    },
    divider: linoraColors.border,
  },
  shape: {
    borderRadius: 16,
  },
  typography: {
    fontFamily:
      "Inter, 'Noto Sans Thai', ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
    h1: { fontSize: "1.8rem", lineHeight: 1.18, fontWeight: 800 },
    h2: { fontSize: "1.35rem", lineHeight: 1.25, fontWeight: 800 },
    h3: { fontSize: "1.08rem", lineHeight: 1.3, fontWeight: 700 },
    button: { fontWeight: 700, textTransform: "none" },
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          fontFamily:
            "Inter, 'Noto Sans Thai', ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          minHeight: 48,
          borderRadius: 16,
          boxShadow: "none",
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: 16,
          border: `1px solid ${linoraColors.border}`,
          boxShadow: "0 10px 28px rgba(36, 40, 38, 0.06)",
        },
      },
    },
  },
});
