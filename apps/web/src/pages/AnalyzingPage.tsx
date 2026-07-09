import { CircularProgress, Stack, Typography } from "@mui/material";

export function AnalyzingPage() {
  return (
    <Stack spacing={2} sx={{ alignItems: "center", py: 8, textAlign: "center" }}>
      <CircularProgress />
      <Typography variant="h1">กำลังวิเคราะห์เพจ</Typography>
      <Typography color="text.secondary">
        กำลังดึงข้อมูลโพสต์ ตรวจคอมเมนต์ และเตรียมผลลัพธ์สำหรับ LINE
      </Typography>
    </Stack>
  );
}
