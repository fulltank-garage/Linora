import { Box, Button, Card, CardContent, Stack, Typography } from "@mui/material";
import { Link, useSearchParams } from "react-router-dom";

type LegalPageProps = {
  type: "privacy" | "terms" | "data-deletion";
};

const legalContent = {
  privacy: {
    title: "ความเป็นส่วนตัว",
    subtitle: "Linora ใช้ข้อมูลเท่าที่จำเป็นเพื่อวิเคราะห์เพจ Facebook ที่คุณเลือก",
    items: [
      "Linora อ่านรายชื่อเพจ Facebook ที่คุณจัดการ เพื่อให้คุณเลือกเพจที่ต้องการวิเคราะห์",
      "Linora ใช้ข้อมูลโพสต์ การมีส่วนร่วม และความคิดเห็นที่จำเป็นเพื่อจัดทำรายงานและคำแนะนำสำหรับเพจที่คุณเลือก",
      "เมื่อสร้างคำแนะนำด้วย AI เราอาจส่งเฉพาะข้อมูลที่เกี่ยวข้องจากเพจ เช่น ข้อความโพสต์ ความคิดเห็น และตัวเลขสรุป ไปยังผู้ให้บริการ AI ภายนอกที่ Linora ใช้งาน เพื่อประมวลผลคำแนะนำให้คุณ",
      "Linora เก็บ Page access token ในรูปแบบเข้ารหัส และไม่แสดง token หรือข้อมูลการเชื่อมต่อ Facebook ของคุณในหน้าเว็บ",
      "Linora ไม่ขายข้อมูล และไม่นำข้อมูลจากเพจ Facebook ไปใช้เพื่อโฆษณาหรือวิเคราะห์พฤติกรรมนอกบริการ",
      "คุณสามารถยกเลิกการเชื่อมต่อหรือลบข้อมูลของเพจที่เชื่อมต่อได้ทุกเมื่อ",
    ],
  },
  terms: {
    title: "เงื่อนไข",
    subtitle: "เงื่อนไขการใช้ Linora สำหรับวิเคราะห์เพจ Facebook",
    items: [
      "คุณต้องเป็นผู้ดูแลหรือมีสิทธิ์จัดการเพจที่เลือก",
      "รายงานจาก Linora เป็นคำแนะนำเชิงวิเคราะห์ ผู้ใช้ควรตรวจสอบก่อนนำไปใช้งานจริง",
      "Linora จะไม่โพสต์ ตอบกลับ หรือแก้ไขข้อมูลบนเพจ Facebook โดยอัตโนมัติ",
      "เมื่อคุณใช้คำแนะนำ AI คุณยอมรับว่าข้อมูลที่จำเป็นจากเพจที่เลือกอาจถูกประมวลผลโดยผู้ให้บริการ AI ภายนอกเพื่อสร้างคำแนะนำ",
      "หากมีฟีเจอร์ตอบกลับในอนาคต Linora จะให้ผู้ใช้ยืนยันก่อนส่งข้อความเสมอ",
    ],
  },
  "data-deletion": {
    title: "ลบข้อมูล",
    subtitle: "วิธีลบข้อมูลที่ Linora ใช้สำหรับวิเคราะห์เพจ Facebook",
    items: [
      "เปิด Linora จาก LINE เลือกเพจที่ต้องการ แล้วกดไอคอนการตั้งค่าและปุ่ม ลบข้อมูล Linora",
      "คำขอลบข้อมูลจะลบ Page access token ที่เข้ารหัส รายงาน ข้อมูลเพจ และตัวเลขที่ Linora เก็บไว้สำหรับเพจนั้น",
      "หากไม่สามารถเข้า LINE ได้ โปรดส่งคำขอลบข้อมูลพร้อม LINE user ID หรือชื่อเพจไปที่ support@linora.app",
      "Linora จะไม่เก็บข้อมูลของเพจที่ลบไว้เพื่อใช้วิเคราะห์ต่อ",
    ],
  },
};

export function LegalPage({ type }: LegalPageProps) {
  const content = legalContent[type];
  const [searchParams] = useSearchParams();
  const confirmationCode = searchParams.get("confirmation_code");

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
      {type === "data-deletion" && confirmationCode ? (
        <Card>
          <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
            <Stack spacing={0.5} sx={{ textAlign: "center" }}>
              <Typography color="primary.main" sx={{ fontSize: 15, fontWeight: 900 }}>
                รับคำขอลบข้อมูลแล้ว
              </Typography>
              <Typography color="text.secondary" sx={{ fontSize: 13, lineHeight: 1.5 }}>
                Linora ดำเนินการลบข้อมูลที่เชื่อมโยงกับคำขอนี้เรียบร้อยแล้ว
              </Typography>
              <Typography color="text.secondary" sx={{ fontSize: 12 }}>
                รหัสยืนยัน: {confirmationCode}
              </Typography>
            </Stack>
          </CardContent>
        </Card>
      ) : null}
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
