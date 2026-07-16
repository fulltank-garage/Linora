import { useState } from "react";
import { Alert, Box, Button, Card, CardContent, Checkbox, FormControlLabel, Stack, Typography } from "@mui/material";
import { CheckCircle, Facebook, RadioButtonUnchecked, ShieldOutlined } from "@mui/icons-material";
import { useNavigate } from "react-router-dom";
import type { FacebookPageSummary } from "@linora/shared";
import { InsightCard } from "@linora/ui";
import { ComplianceLinks } from "../components/ComplianceLinks";
import { LoadingDots } from "../components/LoadingDots";

type PageSelectPageProps = {
  selectedPage: FacebookPageSummary | null;
  pages: FacebookPageSummary[];
  onSelectPage: (page: FacebookPageSummary) => void;
  onAuthorize: () => Promise<void>;
};

export function PageSelectPage({
  selectedPage,
  onSelectPage,
  onAuthorize,
  pages,
}: PageSelectPageProps) {
  const navigate = useNavigate();
  const [hasConfirmedPermission, setHasConfirmedPermission] = useState(false);
  const [isAuthorizing, setIsAuthorizing] = useState(false);
  const [error, setError] = useState("");

  async function handleAuthorize() {
    if (!selectedPage || !hasConfirmedPermission) return;
    setError("");
    setIsAuthorizing(true);
    try {
      await onAuthorize();
      navigate("/dashboard");
    } catch {
      setError("ไม่สามารถเชื่อมต่อและวิเคราะห์เพจได้ กรุณาลองอีกครั้ง");
    } finally {
      setIsAuthorizing(false);
    }
  }

  return (
    <>
      <Stack spacing={2} sx={{ pb: 27 }}>
        <Stack spacing={0.5} sx={{ alignItems: "center", textAlign: "center" }}>
          <Typography variant="h1">เลือกเพจ Facebook</Typography>
          <Typography color="text.secondary" sx={{ fontSize: 15, lineHeight: 1.45 }}>
            เลือกเพจที่ต้องการให้ Linora วิเคราะห์
          </Typography>
        </Stack>
        <Card>
          <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
            <Stack spacing={1}>
              <Typography color="text.primary" sx={{ fontSize: 15, fontWeight: 900 }}>
                ใช้ข้อมูลเฉพาะเพจที่คุณเลือก
              </Typography>
              <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.55 }}>
                Linora จะใช้ข้อมูลเพจ การมีส่วนร่วม และความคิดเห็นที่จำเป็นเพื่อจัดทำรายงานเท่านั้น คุณสามารถเปลี่ยนเพจหรือยกเลิกการเชื่อมต่อได้ภายหลัง
              </Typography>
            </Stack>
          </CardContent>
        </Card>
        {error ? <Alert severity="error">{error}</Alert> : null}
        {pages.map((page) => (
          <InsightCard
            action={
              <Button
                fullWidth
                onClick={() => onSelectPage(page)}
                startIcon={
                  selectedPage?.pageId === page.pageId ? (
                    <CheckCircle />
                  ) : (
                    <RadioButtonUnchecked />
                  )
                }
                sx={{
                  ...(selectedPage?.pageId === page.pageId
                    ? {
                        bgcolor: "#1877F2",
                        "&:hover": {
                          bgcolor: "#166FE5",
                        },
                      }
                    : {
                        borderColor: "#1877F2",
                        color: "#1877F2",
                        "&:hover": {
                          bgcolor: "rgba(24, 119, 242, 0.08)",
                          borderColor: "#1877F2",
                        },
                      }),
                }}
                variant={selectedPage?.pageId === page.pageId ? "contained" : "outlined"}
              >
                {selectedPage?.pageId === page.pageId ? "เลือกเพจนี้แล้ว" : "เลือกเพจนี้"}
              </Button>
            }
            helper={`หมวดหมู่: ${page.category}`}
            key={page.pageId}
            title="เพจที่พร้อมเชื่อมต่อ"
            value={
              <Box sx={{ alignItems: "center", display: "flex", gap: 1 }}>
                <Facebook sx={{ color: "#1877F2", flexShrink: 0, fontSize: 34 }} />
                <Box component="span" sx={{ minWidth: 0, overflow: "hidden", textOverflow: "ellipsis" }}>
                  {page.pageName}
                </Box>
              </Box>
            }
          />
        ))}
        <ComplianceLinks />
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
        <Stack spacing={1} sx={{ maxWidth: 430, mx: "auto" }}>
          <Card>
            <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
              <Stack spacing={1.25}>
                <Typography color="text.primary" sx={{ fontSize: 15, fontWeight: 900 }}>
                  ยืนยันก่อนอนุญาต
                </Typography>
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={hasConfirmedPermission}
                      onChange={(event) => setHasConfirmedPermission(event.target.checked)}
                    />
                  }
                  label={
                    <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.45 }}>
                      ฉันเป็นผู้ดูแลหรือมีสิทธิ์จัดการเพจนี้ และอนุญาตให้ Linora วิเคราะห์ข้อมูลของเพจที่เลือก
                    </Typography>
                  }
                  sx={{ alignItems: "flex-start", m: 0 }}
                />
                <Typography color="text.secondary" sx={{ fontSize: 13, lineHeight: 1.5 }}>
                  Linora จะไม่โพสต์ ตอบกลับ หรือแก้ไขเพจที่เลือกโดยอัตโนมัติ
                </Typography>
              </Stack>
            </CardContent>
          </Card>
          <Button
            disabled={!selectedPage || !hasConfirmedPermission || isAuthorizing}
            fullWidth
            onClick={handleAuthorize}
            size="large"
            startIcon={isAuthorizing ? <LoadingDots color="currentColor" size={7} /> : <ShieldOutlined />}
            variant="contained"
          >
            {isAuthorizing ? "กำลังเปิดหน้าวิเคราะห์" : "อนุญาตและเข้าสู่หน้าวิเคราะห์"}
          </Button>
        </Stack>
      </Box>
    </>
  );
}
