import { Box, Button, Card, CardContent, Stack, Typography } from "@mui/material";
import { Link } from "react-router-dom";

type LegalPageProps = {
  type: "privacy" | "terms" | "data-deletion";
};

const legalContent = {
  privacy: {
    title: "ความเป็นส่วนตัว",
    subtitle: "Linora ใช้ข้อมูลเท่าที่จำเป็นเพื่อวิเคราะห์ Facebook Page ที่คุณเลือก",
    items: [
      "Linora อ่านรายการ Page ที่คุณจัดการเพื่อให้คุณเลือก Page ที่ต้องการวิเคราะห์",
      "Linora ใช้ข้อมูล Page, engagement และคอมเมนต์ที่เกี่ยวข้องเพื่อสร้างรายงานและคำแนะนำ",
      "Linora ไม่แสดงหรือเก็บ page access token ในหน้าเว็บ",
      "Linora ไม่ขาย ไม่แชร์ และไม่ใช้ข้อมูล Facebook Page เพื่อโฆษณาหรือ profiling ภายนอกบริการ",
      "คุณสามารถยกเลิกการเชื่อมต่อหรือลบข้อมูล Linora ได้ทุกเมื่อ",
    ],
  },
  terms: {
    title: "เงื่อนไข",
    subtitle: "เงื่อนไขการใช้งาน Linora สำหรับการวิเคราะห์ Facebook Page",
    items: [
      "ผู้ใช้ต้องเป็นผู้ดูแลหรือมีสิทธิ์จัดการ Page ที่เลือก",
      "รายงานจาก Linora เป็นคำแนะนำเชิงวิเคราะห์ ผู้ใช้ควรตรวจสอบก่อนนำไปใช้งานจริง",
      "Linora จะไม่โพสต์ ตอบกลับ หรือแก้ไขข้อมูลบน Facebook Page โดยอัตโนมัติ",
      "หากมีฟีเจอร์ตอบกลับในอนาคต Linora จะให้ผู้ใช้ยืนยันก่อนส่งข้อความเสมอ",
    ],
  },
  "data-deletion": {
    title: "ลบข้อมูล",
    subtitle: "วิธีลบข้อมูลที่ Linora ใช้สำหรับวิเคราะห์ Page",
    items: [
      "กดปุ่ม ลบข้อมูล Linora ใน Dashboard เพื่อจำลองการลบข้อมูลในเดโมนี้",
      "ระบบจะล้าง Page ที่เลือกและสถานะการอนุญาตออกจาก session ปัจจุบัน",
      "สำหรับระบบจริง คำขอลบข้อมูลจะลบรายงาน ข้อมูล Page ที่เชื่อมต่อ และข้อมูลที่ดึงจาก Facebook",
      "ติดต่อ support@linora.app หากต้องการให้ทีมงานช่วยตรวจสอบคำขอลบข้อมูล",
    ],
  },
};

export function LegalPage({ type }: LegalPageProps) {
  const content = legalContent[type];

  return (
    <>
      <Stack spacing={2} sx={{ pb: 11 }}>
      <Stack spacing={0.75} sx={{ alignItems: "center", textAlign: "center" }}>
        <Typography variant="h1">{content.title}</Typography>
        <Typography color="text.secondary" sx={{ fontSize: 15, lineHeight: 1.5 }}>
          {content.subtitle}
        </Typography>
      </Stack>
      <Card>
        <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
          <Stack component="ul" spacing={1.25} sx={{ m: 0, pl: 2.5 }}>
            {content.items.map((item) => (
              <Typography component="li" key={item} sx={{ color: "text.secondary", fontSize: 14, lineHeight: 1.55 }}>
                {item}
              </Typography>
            ))}
          </Stack>
        </CardContent>
      </Card>
      </Stack>
      <Box
        sx={{
          backdropFilter: "blur(18px)",
          background:
            "linear-gradient(to top, rgba(248, 246, 240, 0.96) 0%, rgba(248, 246, 240, 0.78) 62%, rgba(248, 246, 240, 0) 100%)",
          borderTop: 0,
          bottom: 0,
          boxShadow: "none",
          left: 0,
          p: 2,
          pt: 2.5,
          position: "fixed",
          right: 0,
          WebkitBackdropFilter: "blur(18px)",
          zIndex: 20,
        }}
      >
        <Box sx={{ maxWidth: 430, mx: "auto" }}>
          <Button component={Link} fullWidth size="large" to="/" variant="contained">
            กลับหน้าหลัก
          </Button>
        </Box>
      </Box>
    </>
  );
}
