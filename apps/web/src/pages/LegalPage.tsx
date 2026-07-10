import { Box, Button, Card, CardContent, Stack, Typography } from "@mui/material";
import { Link } from "react-router-dom";

type LegalPageProps = {
  type: "privacy" | "terms" | "data-deletion";
};

const legalContent = {
  privacy: {
    title: "ความเป็นส่วนตัว",
    subtitle: "Linora ใช้ข้อมูลเท่าที่จำเป็นเพื่อวิเคราะห์เพจ Facebook ที่คุณเลือก",
    items: [
      "Linora อ่านรายชื่อเพจ Facebook ที่คุณจัดการ เพื่อให้คุณเลือกเพจที่ต้องการวิเคราะห์",
      "Linora ใช้ข้อมูลเพจ การมีส่วนร่วม และความคิดเห็นที่เกี่ยวข้องเพื่อจัดทำรายงานและคำแนะนำ",
      "Linora จะไม่แสดงข้อมูลการเชื่อมต่อ Facebook ของคุณในหน้าเว็บ",
      "Linora ไม่ขาย ไม่แชร์ และไม่นำข้อมูลจากเพจ Facebook ไปใช้เพื่อโฆษณาหรือวิเคราะห์พฤติกรรมนอกบริการ",
      "คุณสามารถยกเลิกการเชื่อมต่อหรือลบข้อมูล Linora ได้ทุกเมื่อ",
    ],
  },
  terms: {
    title: "เงื่อนไข",
    subtitle: "เงื่อนไขการใช้ Linora สำหรับวิเคราะห์เพจ Facebook",
    items: [
      "คุณต้องเป็นผู้ดูแลหรือมีสิทธิ์จัดการเพจที่เลือก",
      "รายงานจาก Linora เป็นคำแนะนำเชิงวิเคราะห์ ผู้ใช้ควรตรวจสอบก่อนนำไปใช้งานจริง",
      "Linora จะไม่โพสต์ ตอบกลับ หรือแก้ไขข้อมูลบนเพจ Facebook โดยอัตโนมัติ",
      "หากมีฟีเจอร์ตอบกลับในอนาคต Linora จะให้ผู้ใช้ยืนยันก่อนส่งข้อความเสมอ",
    ],
  },
  "data-deletion": {
    title: "ลบข้อมูล",
    subtitle: "วิธีลบข้อมูลที่ Linora ใช้สำหรับวิเคราะห์เพจ Facebook",
    items: [
      "กดปุ่ม ลบข้อมูล Linora ในหน้าสรุปเพื่อลบข้อมูลที่เชื่อมต่อไว้",
      "ระบบจะนำเพจที่เลือกและสิทธิ์ที่อนุญาตออกจากเบราว์เซอร์นี้",
      "คำขอลบข้อมูลจะลบรายงาน ข้อมูลเพจที่เชื่อมต่อ และข้อมูลที่ดึงจาก Facebook",
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
          pb: "calc(8px + env(safe-area-inset-bottom, 0px))",
          pt: 0.75,
          px: 2,
          position: "fixed",
          right: 0,
          transform: "translate3d(0, 0, 0)",
          WebkitTransform: "translate3d(0, 0, 0)",
          willChange: "transform",
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
