import { Box, Button, Card, CardContent, Stack, Typography } from "@mui/material";
import { CheckCircle, Facebook } from "@mui/icons-material";
import { useNavigate } from "react-router-dom";
import { ComplianceLinks } from "../components/ComplianceLinks";

type ConnectFacebookPageProps = {
  hasFacebookLogin: boolean;
  onLogin: () => void;
};

export function ConnectFacebookPage({
  hasFacebookLogin,
  onLogin,
}: ConnectFacebookPageProps) {
  const navigate = useNavigate();

  function handleLogin() {
    onLogin();
    navigate("/pages");
  }

  return (
    <Stack
      sx={{
        minHeight: "calc(100vh - 160px)",
        pb: 1,
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
            เริ่มวิเคราะห์ Facebook Page ของคุณได้ในไม่กี่ขั้นตอน
          </Typography>
        </Stack>
        <Button
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
          {hasFacebookLogin ? "ไปเลือก Page" : "เข้าสู่ระบบด้วย Facebook"}
        </Button>
      </Stack>
      <Stack spacing={1.25} sx={{ flex: "0 0 auto", width: "100%" }}>
        <Card sx={{ width: "100%" }}>
          <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
            <Stack spacing={1}>
              <Typography color="text.primary" sx={{ fontSize: 15, fontWeight: 900 }}>
                Linora ขอใช้ Facebook Login เพื่ออะไร
              </Typography>
              <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.55 }}>
                เราใช้สิทธิ์เพื่ออ่านรายการ Page ที่คุณจัดการ และดึงข้อมูลเชิงสรุปของ Page ที่คุณเลือกสำหรับทำรายงานเท่านั้น
              </Typography>
              <Typography color="text.secondary" sx={{ fontSize: 13, lineHeight: 1.5 }}>
                Linora ไม่ขอรหัสผ่าน Facebook, ไม่แสดง access token ในหน้าเว็บ และไม่โพสต์หรือตอบกลับแทนคุณโดยอัตโนมัติ
              </Typography>
            </Stack>
          </CardContent>
        </Card>
        <ComplianceLinks />
      </Stack>
    </Stack>
  );
}
