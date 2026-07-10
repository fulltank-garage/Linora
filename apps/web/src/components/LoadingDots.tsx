import { Box } from "@mui/material";

export function LoadingDots() {
  return (
    <Box
      aria-label="กำลังโหลด"
      sx={{
        display: "flex",
        gap: 0.75,
        "@keyframes linora-loading-dot-bounce": {
          "0%, 60%, 100%": {
            opacity: 0.45,
            transform: "translateY(0)",
          },
          "30%": {
            opacity: 1,
            transform: "translateY(-8px)",
          },
        },
      }}
    >
      {[0, 1, 2].map((index) => (
        <Box
          component="span"
          key={index}
          sx={{
            animation: "linora-loading-dot-bounce 900ms ease-in-out infinite",
            animationDelay: `${index * 140}ms`,
            bgcolor: "primary.main",
            borderRadius: "50%",
            height: 10,
            width: 10,
          }}
        />
      ))}
    </Box>
  );
}
