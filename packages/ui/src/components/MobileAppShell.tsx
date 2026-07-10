import type { ReactNode } from "react";
import { AppBar, Box, Container, Toolbar } from "@mui/material";

type MobileAppShellProps = {
  title?: string;
  children: ReactNode;
};

export function MobileAppShell({
  title = "Linora",
  children,
}: MobileAppShellProps) {
  return (
    <Box sx={{ minHeight: "100vh", bgcolor: "background.default" }}>
      <AppBar
        color="inherit"
        elevation={0}
        position="fixed"
        sx={{
          bgcolor: "transparent",
          borderBottom: 0,
          isolation: "isolate",
          overflow: "visible",
          left: 0,
          position: "fixed",
          right: 0,
          top: 0,
          zIndex: (theme) => theme.zIndex.appBar,
          "&::after": {
            backdropFilter: "blur(18px)",
            backgroundColor: "rgba(255, 255, 255, 0.82)",
            borderRadius: "0 0 50% 50% / 0 0 100% 100%",
            content: '\"\"',
            height: 72,
            left: "50%",
            pointerEvents: "none",
            position: "absolute",
            top: -10,
            transform: "translateX(-50%)",
            width: "min(470px, 108vw)",
            WebkitBackdropFilter: "blur(18px)",
            zIndex: 0,
          },
        }}
      >
        <Toolbar
          sx={{
            justifyContent: "center",
            maxWidth: 430,
            minHeight: 60,
            mx: "auto",
            px: 0,
            position: "relative",
            zIndex: 2,
            width: "100%",
          }}
        >
          <Box
            alt={title}
            component="img"
            src="/linora-logo-tight.png"
            sx={{
              display: "block",
              height: 40,
              left: "50%",
              maxWidth: 210,
              objectFit: "contain",
              position: "absolute",
              transform: "translateX(-50%)",
              width: "auto",
            }}
          />
        </Toolbar>
      </AppBar>
      <Container
        maxWidth={false}
        sx={{
          maxWidth: 430,
          px: 2,
          pt: 11,
          pb: 4,
        }}
      >
        {children}
      </Container>
    </Box>
  );
}
