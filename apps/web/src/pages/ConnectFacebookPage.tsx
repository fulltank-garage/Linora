import { useState } from "react";
import { Alert, Box, Button, Stack, Typography } from "@mui/material";
import { CheckCircle, Facebook } from "@mui/icons-material";
import { useNavigate } from "react-router-dom";
import { ComplianceLinks } from "../components/ComplianceLinks";
import { LoadingDots } from "../components/LoadingDots";

type ConnectFacebookPageProps = {
  hasFacebookLogin: boolean;
  isLoading?: boolean;
  loginError?: string | null;
  onLogin: () => void;
};

export function ConnectFacebookPage({
  hasFacebookLogin,
  isLoading = false,
  loginError,
  onLogin,
}: ConnectFacebookPageProps) {
  const navigate = useNavigate();
  const [isRedirecting, setIsRedirecting] = useState(false);

  function handleLogin() {
    if (hasFacebookLogin) {
      navigate("/pages");
      return;
    }
    setIsRedirecting(true);
    window.setTimeout(onLogin, 650);
  }

  return (
    <Stack
      sx={{
        mb: -4,
        minHeight: "calc(100dvh - 88px)",
        pb: 0,
        textAlign: "center",
      }}
    >
      <Stack
        spacing={2.25}
        sx={{
          alignItems: "center",
          flex: "1 1 auto",
          justifyContent: "center",
          minHeight: 0,
          transform: "translateY(-28px)",
          width: "100%",
        }}
      >
        <Box
          alt="Linora"
          component="img"
          src="/linora-icon.png"
          sx={{
            display: "block",
            height: 118,
            objectFit: "contain",
            width: 118,
          }}
        />
        <Stack spacing={0.75}>
          <Typography variant="h1">ยินดีต้อนรับสู่ Linora</Typography>
          <Typography color="text.secondary" sx={{ fontSize: 15, lineHeight: 1.5 }}>
            เริ่มวิเคราะห์เพจ Facebook ของคุณได้ในไม่กี่ขั้นตอน
          </Typography>
        </Stack>
        {loginError && !isRedirecting ? <Alert severity="error">{loginError}</Alert> : null}
        <Button
          disabled={isLoading || isRedirecting}
          onClick={handleLogin}
          size="large"
          startIcon={hasFacebookLogin ? <CheckCircle /> : <Facebook />}
          sx={{
            bgcolor: "#1877F2",
            boxShadow: "0 10px 22px rgba(24, 119, 242, 0.22)",
            maxWidth: 320,
            minWidth: 260,
            width: "82%",
            "&:hover": {
              bgcolor: "#166FE5",
              boxShadow: "0 12px 26px rgba(24, 119, 242, 0.28)",
            },
          }}
          variant="contained"
        >
          {isLoading || isRedirecting
            ? "กำลังเปิด Facebook"
            : hasFacebookLogin
              ? "ไปเลือก Page"
              : "เข้าสู่ระบบด้วย Facebook"}
        </Button>
      </Stack>
      <Stack
        spacing={1.25}
        sx={{
          bgcolor: "background.paper",
          borderRadius: "50% 50% 0 0 / 28px 28px 0 0",
          flex: "0 0 auto",
          mx: -2,
          pb: "calc(12px + env(safe-area-inset-bottom, 0px))",
          pt: 2.5,
          px: 2,
          width: "calc(100% + 32px)",
        }}
      >
        <Stack spacing={1}>
          <Typography color="text.primary" sx={{ fontSize: 15, fontWeight: 900 }}>
            ทำไม Linora จึงขอเชื่อมต่อ Facebook
          </Typography>
          <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.55 }}>
            เราจะอ่านรายชื่อเพจ Facebook ที่คุณจัดการ และใช้ข้อมูลจากเพจที่คุณเลือกเพื่อจัดทำรายงานเท่านั้น
          </Typography>
          <Typography color="text.secondary" sx={{ fontSize: 13, lineHeight: 1.5 }}>
            Linora จะไม่ขอรหัสผ่าน Facebook ไม่แสดงข้อมูลการเชื่อมต่อในหน้าเว็บ และไม่โพสต์หรือตอบกลับแทนคุณโดยอัตโนมัติ
          </Typography>
        </Stack>
        <ComplianceLinks />
      </Stack>
      {isRedirecting ? (
        <Box
          aria-live="polite"
          role="status"
          sx={{
            alignItems: "center",
            backdropFilter: "blur(12px)",
            bgcolor: "rgba(248, 246, 240, 0.96)",
            display: "flex",
            inset: 0,
            justifyContent: "center",
            position: "fixed",
            WebkitBackdropFilter: "blur(12px)",
            zIndex: 100,
          }}
        >
          <Stack spacing={1.25} sx={{ alignItems: "center" }}>
            <LoadingDots />
            <Typography sx={{ fontSize: 19, fontWeight: 900 }}>กรุณารอสักครู่</Typography>
            <Typography color="text.secondary" sx={{ fontSize: 14 }}>
              กำลังเปิดหน้าอนุญาตจาก Facebook
            </Typography>
          </Stack>
        </Box>
      ) : null}
    </Stack>
  );
}
