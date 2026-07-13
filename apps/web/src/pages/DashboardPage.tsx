import { useEffect, useState } from "react";
import { Alert, Box, Button, Card, CardContent, Chip, CircularProgress, Drawer, IconButton, Stack, Tooltip, Typography } from "@mui/material";
import {
  AccessTime,
  AutoAwesome,
  BarChart,
  ChatBubbleOutlined,
  CheckCircle,
  DeleteOutlined,
  Facebook,
  LinkOff,
  SecurityOutlined,
  SettingsOutlined,
  ThumbUpOutlined,
  VisibilityOutlined,
  WorkspacePremium,
} from "@mui/icons-material";
import { LoadingDots } from "../components/LoadingDots";
import { CountUp } from "../components/CountUp";
import { useNavigate } from "react-router-dom";
import type { AnalysisReport, FacebookPageSummary, WeeklyReport } from "@linora/shared";
import { ComplianceLinks } from "../components/ComplianceLinks";

type DashboardPageProps = {
  onAnalyze: () => Promise<void>;
  onDeleteData: () => Promise<void>;
  onDisconnect: () => Promise<void>;
  page: FacebookPageSummary;
  report: AnalysisReport;
  weeklyReport: WeeklyReport | null;
};

function getHealthLabel(score: number) {
  if (score >= 80) return "ดีมาก";
  if (score >= 65) return "ดี";
  if (score >= 45) return "ควรปรับปรุง";
  return "ต้องดูแลเร่งด่วน";
}

function formatMetric(value: number | undefined) {
  if (!value) return "-";
  return new Intl.NumberFormat("th-TH").format(value);
}

function formatWeeklyDateRange(startDate: string, endDate: string) {
  const formatter = new Intl.DateTimeFormat("th-TH", {
    day: "numeric",
    month: "short",
    year: "numeric",
    timeZone: "Asia/Bangkok",
  });
  return `${formatter.format(new Date(`${startDate}T00:00:00+07:00`))} - ${formatter.format(new Date(`${endDate}T00:00:00+07:00`))}`;
}

export function DashboardPage({ onAnalyze, onDeleteData, onDisconnect, page, report, weeklyReport }: DashboardPageProps) {
  const navigate = useNavigate();
  const [isManagementOpen, setIsManagementOpen] = useState(false);
  const [managementAction, setManagementAction] = useState<"delete" | "disconnect" | null>(null);
  const [managementError, setManagementError] = useState("");
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [analysisError, setAnalysisError] = useState("");
  const [scoreAnimationKey, setScoreAnimationKey] = useState(0);
  const [visibleScore, setVisibleScore] = useState(0);
  const [isScoreAnimating, setIsScoreAnimating] = useState(false);
  const healthLabel = getHealthLabel(report.healthScore);
  const postingTimeInsight = report.postingTimeInsight;
  const hasPostingTimeData = Boolean(postingTimeInsight && postingTimeInsight.basedOnPosts > 0);
  const bestPostingTime = postingTimeInsight?.bestTime || "ยังไม่มีข้อมูล";
  const importantComment = report.importantComments[0];
  const postingDays = postingTimeInsight?.days ?? ["จ.", "อ.", "พ.", "พฤ.", "ศ.", "ส.", "อา."].map((day) => ({
    averageEngagement: 0,
    day,
    postCount: 0,
  }));
  const maxAverageEngagement = Math.max(...postingDays.map((day) => day.averageEngagement), 1);
  const postingBars = postingDays.map((day) => ({
    ...day,
    active: day.day === postingTimeInsight?.bestDay && day.postCount > 0,
    value: day.postCount > 0 ? Math.max(16, Math.round((day.averageEngagement / maxAverageEngagement) * 62)) : 5,
  }));
  const weeklyMetrics = weeklyReport?.metrics;
  const latestReportMetrics = [
    { icon: <VisibilityOutlined />, label: "การเข้าถึง", value: formatMetric(weeklyMetrics?.reach) },
    { icon: <ThumbUpOutlined />, label: "การมีส่วนร่วม", value: formatMetric(weeklyMetrics?.engagements) },
  ];

  useEffect(() => {
    setIsScoreAnimating(false);
    setVisibleScore(0);
    let frame = window.requestAnimationFrame(() => {
      frame = window.requestAnimationFrame(() => {
        setIsScoreAnimating(true);
        setVisibleScore(report.healthScore);
      });
    });

    return () => window.cancelAnimationFrame(frame);
  }, [report.healthScore, scoreAnimationKey]);

  async function handleManagementAction(action: "delete" | "disconnect") {
    if (action === "delete" && !window.confirm("ลบข้อมูล Page, รายงาน และตัวเลขวิเคราะห์ทั้งหมดของ Linora ใช่หรือไม่?")) {
      return;
    }
    setManagementAction(action);
    setManagementError("");
    try {
      if (action === "disconnect") {
        await onDisconnect();
      } else {
        await onDeleteData();
      }
      setIsManagementOpen(false);
    } catch {
      setManagementError("ดำเนินการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง");
    } finally {
      setManagementAction(null);
    }
  }

  async function handlePageAnalysis() {
    setAnalysisError("");
    setIsAnalyzing(true);
    try {
      await onAnalyze();
      setScoreAnimationKey((key) => key + 1);
    } catch {
      setAnalysisError("ไม่สามารถวิเคราะห์เพจได้ กรุณาลองใหม่อีกครั้ง");
    } finally {
      setIsAnalyzing(false);
    }
  }

  return (
    <>
      <Stack spacing={2} sx={{ pb: 19 }}>
        <Card>
          <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
            <Stack spacing={1.25}>
              <Box
                sx={{
                  alignItems: "center",
                  display: "flex",
                  gap: 1.25,
                  justifyContent: "space-between",
                  width: "100%",
                }}
              >
                <Box
                  sx={{
                    alignItems: "center",
                    display: "flex",
                    flex: "1 1 auto",
                    gap: 1,
                    minWidth: 0,
                  }}
                >
                  <Facebook sx={{ color: "#1877F2", flexShrink: 0, fontSize: 30 }} />
                  <Typography
                    component="span"
                    sx={{
                      color: "text.primary",
                      fontSize: 22,
                      fontWeight: 800,
                      lineHeight: 1.15,
                      minWidth: 0,
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      whiteSpace: "nowrap",
                    }}
                  >
                    {page.pageName}
                  </Typography>
                </Box>
                <Chip
                  icon={<CheckCircle />}
                  label="อนุญาตแล้ว"
                  size="small"
                  sx={{
                    bgcolor: "#12B76A",
                    color: "#fff",
                    flexShrink: 0,
                    fontWeight: 800,
                    px: 0.75,
                    "& .MuiChip-icon": {
                      color: "#fff",
                      fontSize: 16,
                      ml: 0.5,
                    },
                    "& .MuiChip-label": {
                      px: 1,
                    },
                  }}
                />
              </Box>
              <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.55 }}>
                {`กำลังวิเคราะห์เพจหมวด ${page.category}`}
              </Typography>
              <Box
                sx={{
                  alignItems: "center",
                  bgcolor: "rgba(24, 119, 242, 0.08)",
                  borderRadius: 0.5,
                  display: "flex",
                  gap: 1,
                  px: 1.5,
                  py: 1.1,
                }}
              >
                <SecurityOutlined sx={{ color: "#1877F2", flexShrink: 0, fontSize: 21 }} />
                <Typography color="text.secondary" sx={{ fontSize: 13, lineHeight: 1.45 }}>
                  Linora ใช้ข้อมูลจากเพจที่เลือกเท่านั้น และจะไม่โพสต์แทนคุณโดยอัตโนมัติ
                </Typography>
              </Box>
            </Stack>
          </CardContent>
        </Card>

        <Card sx={{ position: "relative" }}>
          <WorkspacePremium
            sx={{
              color: "#D9A62A",
              fontSize: 34,
              position: "absolute",
              right: 20,
              top: 14,
              zIndex: 1,
            }}
          />
          <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
            <Stack spacing={1}>
              <Typography color="text.secondary" sx={{ fontSize: 13, fontWeight: 800, pr: 5 }}>
                ภาพรวมคุณภาพเพจ
              </Typography>
              <Box
                sx={{
                  alignItems: "center",
                  display: "flex",
                  gap: 1.75,
                  mt: "10px !important",
                }}
              >
                <Box
                  sx={{
                    alignItems: "center",
                    display: "grid",
                    flexShrink: 0,
                    height: 96,
                    justifyItems: "center",
                    placeItems: "center",
                    position: "relative",
                    width: 96,
                  }}
                >
                  <CircularProgress
                    size={96}
                    thickness={4}
                    value={100}
                    variant="determinate"
                    sx={{
                      color: "rgba(15, 148, 117, 0.12)",
                      gridArea: "1 / 1",
                    }}
                  />
                  <CircularProgress
                    size={96}
                    thickness={4}
                    value={visibleScore}
                    variant="determinate"
                    sx={{
                      color: "primary.main",
                      gridArea: "1 / 1",
                      transform: "rotate(-96deg) !important",
                      "& .MuiCircularProgress-circle": {
                        transition: isScoreAnimating
                          ? "stroke-dashoffset 1.2s cubic-bezier(0.22, 1, 0.36, 1)"
                          : "none",
                      },
                    }}
                  />
                  <Stack spacing={0} sx={{ alignItems: "center", gridArea: "1 / 1" }}>
                    <Typography color="primary.main" sx={{ fontSize: 30, fontWeight: 900, lineHeight: 1 }}>
                      <CountUp
                        duration={1.2}
                        from={0}
                        key={`${report.healthScore}-${scoreAnimationKey}`}
                        to={report.healthScore}
                      />
                    </Typography>
                    <Typography color="text.secondary" sx={{ fontSize: 13, fontWeight: 700 }}>
                      /100
                    </Typography>
                  </Stack>
                </Box>
                <Stack spacing={0.5} sx={{ minWidth: 0, pt: 0.25 }}>
                  <Typography color="primary.main" sx={{ fontSize: 20, fontWeight: 900, lineHeight: 1.15 }}>
                    {healthLabel}
                  </Typography>
                  <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.45 }}>
                    {report.summary}
                  </Typography>
                  <Typography color="primary.main" sx={{ fontSize: 13, fontWeight: 800 }}>
                    เพิ่มขึ้น ▲ 8 จากสัปดาห์ที่แล้ว
                  </Typography>
                </Stack>
              </Box>
            </Stack>
          </CardContent>
        </Card>

        <Card sx={{ position: "relative" }}>
          <ChatBubbleOutlined
            sx={{
              color: "#D9A62A",
              fontSize: 38,
              position: "absolute",
              right: 18,
              top: 14,
              zIndex: 1,
            }}
          />
          <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
            <Stack spacing={1.75}>
              <Box sx={{ pr: 6 }}>
                <Typography color="text.secondary" sx={{ fontSize: 13, fontWeight: 800 }}>
                  คอมเมนต์สำคัญ
                </Typography>
                <Typography
                  color="text.primary"
                  sx={{ fontSize: 18, fontWeight: 900, lineHeight: 1.25, mt: "10px !important" }}
                >
                  {importantComment?.message}
                </Typography>
              </Box>
              <Box
                sx={{
                  bgcolor: "rgba(217, 166, 42, 0.1)",
                  borderRadius: 0.5,
                  px: 1.5,
                  py: 1.25,
                }}
              >
                <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.55 }}>
                  {importantComment?.reason}
                </Typography>
              </Box>
            </Stack>
          </CardContent>
        </Card>

        <Card sx={{ position: "relative" }}>
          <WorkspacePremium
            sx={{
              color: "#D9A62A",
              fontSize: 34,
              position: "absolute",
              right: 20,
              top: 14,
              zIndex: 1,
            }}
          />
          <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
            <Stack spacing={1.25}>
              <Typography color="text.secondary" sx={{ fontSize: 13, fontWeight: 800, pr: 5 }}>
                เวลาโพสต์ที่ดีที่สุด
              </Typography>
              <Box
                sx={{
                  alignItems: "center",
                  display: "flex",
                  gap: 0.75,
                  whiteSpace: "nowrap",
                }}
              >
                <AccessTime sx={{ color: "#D9A62A", flexShrink: 0, fontSize: 27 }} />
                <Typography
                  color="text.primary"
                  sx={{ fontSize: 25, fontWeight: 900, lineHeight: 1, whiteSpace: "nowrap" }}
                >
                  {bestPostingTime}
                </Typography>
              </Box>
              <Typography color="text.secondary" sx={{ fontSize: 12, fontWeight: 700, lineHeight: 1.35 }}>
                {hasPostingTimeData
                  ? `อ้างอิงจากโพสต์ล่าสุด ${postingTimeInsight?.basedOnPosts} โพสต์`
                  : "วิเคราะห์เพจอีกครั้งเพื่อคำนวณจากโพสต์จริง"}
              </Typography>
              <Box
                sx={{
                  alignItems: "end",
                  display: "grid",
                  gap: 0.8,
                  gridTemplateColumns: "repeat(7, 1fr)",
                  minWidth: 0,
                  pt: 2.5,
                }}
              >
                {postingBars.map((bar) => (
                  <Stack key={bar.day} spacing={0.65} sx={{ alignItems: "center" }}>
                   <Box
                     sx={{
                       alignItems: "end",
                       display: "flex",
                       height: 76,
                       justifyContent: "center",
                       position: "relative",
                       width: "100%",
                     }}
                   >
                    {bar.active ? (
                      <Typography
                        sx={{
                          color: "#E53935",
                          fontSize: 12,
                          fontWeight: 900,
                          left: "50%",
                          lineHeight: 1,
                          position: "absolute",
                          top: -9,
                          transform: "translateX(-50%)",
                          whiteSpace: "nowrap",
                        }}
                      >
                        ดีที่สุด
                      </Typography>
                    ) : null}
                    <Box
                      sx={{
                        background: bar.active
                          ? "linear-gradient(180deg, #0F9475 0%, #0B7D65 100%)"
                          : "linear-gradient(180deg, rgba(15, 148, 117, 0.42) 0%, rgba(15, 148, 117, 0.22) 100%)",
                        borderRadius: "6px 6px 0 0",
                        height: bar.value,
                        width: "100%",
                      }}
                    />
                  </Box>
                  <Typography color="text.secondary" sx={{ fontSize: 12, fontWeight: 700 }}>
                    {bar.day}
                  </Typography>
                </Stack>
              ))}
              </Box>
            </Stack>
          </CardContent>
        </Card>

        <Card sx={{ position: "relative" }}>
          <BarChart
            sx={{
              color: "#D9A62A",
              fontSize: 42,
              position: "absolute",
              right: 18,
              top: 14,
              zIndex: 1,
            }}
          />
          <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
            <Stack spacing={1.35}>
              <Box sx={{ pr: 6 }}>
                <Typography color="text.secondary" sx={{ fontSize: 13, fontWeight: 800 }}>
                  รายงานล่าสุด
                </Typography>
                <Typography color="text.primary" sx={{ fontSize: 18, fontWeight: 900, lineHeight: 1.25, mt: 0.5 }}>
                  รายงาน 7 วันล่าสุด
                </Typography>
                <Typography color="text.secondary" sx={{ fontSize: 14, fontWeight: 600 }}>
                  {weeklyReport ? formatWeeklyDateRange(weeklyReport.startDate, weeklyReport.endDate) : "กำลังโหลดช่วงรายงาน"}
                </Typography>
                <Typography color="text.secondary" sx={{ fontSize: 12, fontWeight: 700 }}>
                  {weeklyReport ? `มีข้อมูล ${weeklyReport.daysWithData}/7 วัน` : "ยังไม่มีข้อมูลรายวัน"}
                </Typography>
              </Box>
              <Stack spacing={1.15}>
                {latestReportMetrics.map((metric) => (
                  <Box
                    key={metric.label}
                    sx={{
                      alignItems: "center",
                      display: "grid",
                      gap: 1,
                      gridTemplateColumns: "24px 1fr auto",
                    }}
                  >
                    <Box
                      sx={{
                        color: "text.secondary",
                        display: "grid",
                        placeItems: "center",
                        "& svg": { fontSize: 20 },
                      }}
                    >
                      {metric.icon}
                    </Box>
                    <Typography color="text.secondary" sx={{ fontSize: 14, fontWeight: 700 }}>
                      {metric.label}
                    </Typography>
                    <Typography color="text.primary" sx={{ fontSize: 17, fontWeight: 900 }}>
                      {metric.value}
                    </Typography>
                  </Box>
                ))}
              </Stack>
            </Stack>
          </CardContent>
        </Card>

        <Card sx={{ position: "relative" }}>
          <AutoAwesome
            sx={{
              color: "#D9A62A",
              fontSize: 34,
              position: "absolute",
              right: 20,
              top: 14,
              zIndex: 1,
            }}
          />
          <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
            <Stack spacing={1.75}>
              <Box sx={{ alignItems: "center", display: "flex", gap: 1, pr: 5 }}>
                <Stack spacing={0.55} sx={{ minWidth: 0 }}>
                  <Typography color="text.secondary" sx={{ fontSize: 13, fontWeight: 800 }}>
                    คำแนะนำจาก Linora
                  </Typography>
                  <Typography color="text.primary" sx={{ fontSize: 18, fontWeight: 900, lineHeight: 1.25, mt: "10px !important" }}>
                    {report.topPosts[0]?.reason}
                  </Typography>
                </Stack>
              </Box>
              <Box
                sx={{
                  bgcolor: "rgba(15, 148, 117, 0.07)",
                  borderRadius: 0.5,
                  px: 1.5,
                  py: 1.25,
                }}
              >
                <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.55 }}>
                  {report.topPosts[0]?.recommendation}
                </Typography>
              </Box>
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
        <Stack spacing={1} sx={{ maxWidth: 430, mx: "auto" }}>
          {analysisError ? <Alert severity="error">{analysisError}</Alert> : null}
          <Button
            disabled={isAnalyzing}
            fullWidth
            onClick={() => void handlePageAnalysis()}
            size="large"
            startIcon={isAnalyzing ? <LoadingDots color="currentColor" size={7} /> : <AutoAwesome />}
            variant="contained"
          >
            {isAnalyzing ? "กรุณารอสักครู่" : "เริ่มวิเคราะห์เพจ"}
          </Button>
          <Stack direction="row" spacing={1}>
            <Button
              fullWidth
              onClick={() => navigate("/pages")}
              size="large"
              startIcon={<Facebook />}
              sx={{
                bgcolor: "rgba(255, 255, 255, 0.42)",
                borderColor: "#1877F2",
                color: "#1877F2",
                "&:hover": {
                  bgcolor: "rgba(24, 119, 242, 0.08)",
                  borderColor: "#1877F2",
                },
              }}
              variant="outlined"
            >
              เปลี่ยนเพจ Facebook
            </Button>
            <Tooltip title="จัดการข้อมูลและสิทธิ์">
              <IconButton
                aria-label="จัดการข้อมูลและสิทธิ์"
                onClick={() => setIsManagementOpen(true)}
                sx={{
                  border: "1px solid",
                  borderColor: "#1877F2",
                  borderRadius: 1,
                  color: "#1877F2",
                  flex: "0 0 52px",
                  height: 52,
                  width: 52,
                  "&:hover": { bgcolor: "rgba(24, 119, 242, 0.08)" },
                }}
              >
                <SettingsOutlined />
              </IconButton>
            </Tooltip>
          </Stack>
        </Stack>
      </Box>
      <Drawer
        anchor="bottom"
        onClose={() => setIsManagementOpen(false)}
        open={isManagementOpen}
        slotProps={{
          paper: {
            sx: {
              borderRadius: "50% 50% 0 0 / 28px 28px 0 0",
            maxWidth: 430,
            mx: "auto",
            pb: "calc(12px + env(safe-area-inset-bottom, 0px))",
            px: 2,
              pt: 2.5,
              width: "100%",
            },
          },
        }}
      >
        <Box sx={{ bgcolor: "divider", borderRadius: 1, height: 4, mx: "auto", mb: 1.5, width: 40 }} />
        <Stack spacing={1.25}>
          <Typography align="center" color="text.primary" sx={{ fontSize: 16, fontWeight: 900 }}>
            การจัดการข้อมูลและสิทธิ์
          </Typography>
          <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.55 }}>
            คุณสามารถยกเลิกการเชื่อมต่อหรือลบข้อมูล Linora ได้ทุกเมื่อ การลบข้อมูลจะนำเพจที่เลือกและสิทธิ์ที่อนุญาตออกจากเบราว์เซอร์นี้
          </Typography>
          {managementError ? <Alert severity="error">{managementError}</Alert> : null}
          <Button
            disabled={managementAction !== null}
            fullWidth
            onClick={() => void handleManagementAction("disconnect")}
            startIcon={<LinkOff />}
            sx={{
              borderColor: "#1877F2",
              color: "#1877F2",
              "&:hover": { bgcolor: "rgba(24, 119, 242, 0.08)", borderColor: "#1877F2" },
            }}
            variant="outlined"
          >
            ยกเลิกการเชื่อมต่อ Facebook
          </Button>
          <Button
            color="error"
            disabled={managementAction !== null}
            fullWidth
            onClick={() => void handleManagementAction("delete")}
            startIcon={<DeleteOutlined />}
            variant="contained"
          >
            ลบข้อมูล Linora
          </Button>
          <ComplianceLinks />
        </Stack>
      </Drawer>
    </>
  );
}
