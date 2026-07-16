import { lazy, Suspense, useEffect, useRef, useState } from "react";
import { Box, Paper, Typography } from "@mui/material";
import { BrowserRouter, Navigate, Route, Routes, useLocation, useNavigate } from "react-router-dom";
import { MobileAppShell } from "@linora/ui";
import type { AnalysisReport, AnalysisStatus, FacebookPageSummary, WeeklyReport } from "@linora/shared";
import {
  activateConnectRichMenu,
  activateDashboardRichMenu,
  completeFacebookLogin,
  connectFacebookPage,
  deleteFacebookPageData,
  disconnectFacebookPage,
  getConnectedFacebookPages,
  getSavedFacebookDashboard,
  getWeeklyFacebookReport,
  selectConnectedFacebookPage,
  startFacebookLogin,
} from "./api/client";
import { LoadingDots } from "./components/LoadingDots";
import { initializeLineIdentity } from "./lib/line";

const facebookTransitionDelay = 650;
const AnalyzingPage = lazy(async () => ({ default: (await import("./pages/AnalyzingPage")).AnalyzingPage }));
const ConnectFacebookPage = lazy(async () => ({ default: (await import("./pages/ConnectFacebookPage")).ConnectFacebookPage }));
const DashboardPage = lazy(async () => ({ default: (await import("./pages/DashboardPage")).DashboardPage }));
const LegalPage = lazy(async () => ({ default: (await import("./pages/LegalPage")).LegalPage }));
const PageSelectPage = lazy(async () => ({ default: (await import("./pages/PageSelectPage")).PageSelectPage }));

function RouteLoading() {
  return (
    <Box
      sx={{
        alignItems: "center",
        display: "flex",
        flexDirection: "column",
        gap: 1.5,
        justifyContent: "center",
        minHeight: "calc(100dvh - 150px)",
        textAlign: "center",
      }}
    >
      <LoadingDots />
      <Typography color="text.secondary" sx={{ fontSize: 14 }}>
        กำลังเปิดหน้าให้คุณ
      </Typography>
    </Box>
  );
}

function DashboardAnalysisPending({ page, status }: { page: FacebookPageSummary; status: AnalysisStatus | null }) {
  const isFailed = status?.state === "failed";
  return (
    <Box
      sx={{
        alignItems: "center",
        display: "flex",
        flexDirection: "column",
        gap: 1.5,
        justifyContent: "center",
        minHeight: "calc(100dvh - 150px)",
        px: 3,
        textAlign: "center",
      }}
    >
      <LoadingDots color={isFailed ? "error.main" : "primary.main"} />
      <Typography sx={{ fontSize: 19, fontWeight: 900 }}>
        {isFailed ? "ยังวิเคราะห์เพจไม่สำเร็จ" : "กำลังเตรียมรายงานของเพจ"}
      </Typography>
      <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.55 }}>
        {isFailed ? `Linora จะลองอัปเดตข้อมูลของ ${page.pageName} อีกครั้งในไม่ช้า` : `กำลังอ่านข้อมูลล่าสุดของ ${page.pageName} เพื่อจัดทำรายงานให้คุณ`}
      </Typography>
    </Box>
  );
}

function AppRoutes() {
  const location = useLocation();
  const navigate = useNavigate();
  const facebookHandoff = new URLSearchParams(location.search).get("facebook_connect");
  const [latestReport, setLatestReport] = useState<AnalysisReport | null>(null);
  const [analysisStatus, setAnalysisStatus] = useState<AnalysisStatus | null>(null);
  const [weeklyReport, setWeeklyReport] = useState<WeeklyReport | null>(null);
  const [hasFacebookLogin, setHasFacebookLogin] = useState(false);
  const [facebookPages, setFacebookPages] = useState<FacebookPageSummary[]>([]);
  const [facebookLoginError, setFacebookLoginError] = useState<string | null>(null);
  const [facebookHandoffCode, setFacebookHandoffCode] = useState<string | null>(null);
  const [isCompletingFacebookLogin, setIsCompletingFacebookLogin] = useState(false);
  const [isLineIdentityReady, setIsLineIdentityReady] = useState(false);
  const completedFacebookHandoff = useRef<string | null>(null);
  const [selectedPage, setSelectedPage] = useState<FacebookPageSummary | null>(null);
  const [hasPagePermission, setHasPagePermission] = useState(false);

  const canViewDashboard = hasFacebookLogin && selectedPage !== null && hasPagePermission;

  function clearFacebookSession() {
    setHasFacebookLogin(false);
    setFacebookPages([]);
    setFacebookHandoffCode(null);
    setLatestReport(null);
    setAnalysisStatus(null);
    setWeeklyReport(null);
    setSelectedPage(null);
    setHasPagePermission(false);
  }

  useEffect(() => {
    let isActive = true;
    void initializeLineIdentity()
      .then(async (ready) => {
        if (!ready || !isActive) return;
        const [pages, dashboard] = await Promise.all([getConnectedFacebookPages(), getSavedFacebookDashboard()]);
        if (!isActive) return;
        setFacebookPages(pages);
        setHasFacebookLogin(pages.length > 0);
        if (dashboard) {
          setSelectedPage(dashboard.page);
          setLatestReport(dashboard.report ?? null);
          setAnalysisStatus(dashboard.analysisStatus);
          setWeeklyReport(dashboard.weeklyReport);
          setHasPagePermission(true);
        }
      })
      .catch(() => {
        if (isActive) clearFacebookSession();
      })
      .finally(() => {
        if (isActive) setIsLineIdentityReady(true);
      });
    return () => {
      isActive = false;
    };
  }, []);

  useEffect(() => {
    if (!isLineIdentityReady || !selectedPage || !hasPagePermission || analysisStatus?.state === "ready") return;
    let isActive = true;
    const refreshDashboard = async () => {
      try {
        const dashboard = await getSavedFacebookDashboard();
        if (!isActive || !dashboard) return;
        setLatestReport(dashboard.report ?? null);
        setAnalysisStatus(dashboard.analysisStatus);
        setWeeklyReport(dashboard.weeklyReport);
      } catch {
        // Keep the most recent report on screen while the background job retries.
      }
    };
    void refreshDashboard();
    const interval = window.setInterval(() => void refreshDashboard(), 2500);
    return () => {
      isActive = false;
      window.clearInterval(interval);
    };
  }, [analysisStatus?.state, hasPagePermission, isLineIdentityReady, selectedPage]);

  useEffect(() => {
	if (!isLineIdentityReady) return;

    const params = new URLSearchParams(location.search);
    const loginError = params.get("facebook_error");
    const handoffCode = params.get("facebook_connect");
    if (loginError) {
      setFacebookLoginError(loginError);
      navigate("/connect-facebook", { replace: true });
      return;
    }
    if (!handoffCode) return;

    // React StrictMode re-runs effects in development. The server handoff is
    // single-use, so never redeem the same OAuth result more than once.
    if (completedFacebookHandoff.current === handoffCode) return;
    completedFacebookHandoff.current = handoffCode;
    const startedAt = Date.now();

    setIsCompletingFacebookLogin(true);
    setFacebookLoginError(null);
    void completeFacebookLogin(handoffCode)
      .then(async (pages) => {
        if (pages.length === 0) {
          setFacebookLoginError("ไม่พบเพจ Facebook ที่คุณมีสิทธิ์จัดการ");
          navigate("/connect-facebook", { replace: true });
          return;
        }

        const remainingDelay = Math.max(0, facebookTransitionDelay - (Date.now() - startedAt));
        if (remainingDelay > 0) {
          await new Promise<void>((resolve) => window.setTimeout(resolve, remainingDelay));
        }

        setFacebookPages(pages);
        setFacebookHandoffCode(handoffCode);
        setHasFacebookLogin(true);
        navigate("/pages", { replace: true });
      })
      .catch(() => {
        setFacebookLoginError("ไม่สามารถรับข้อมูลเพจ Facebook ได้ กรุณาลองใหม่อีกครั้ง");
        navigate("/connect-facebook", { replace: true });
      })
      .finally(() => setIsCompletingFacebookLogin(false));
  }, [isLineIdentityReady, location.search, navigate]);

  async function authorizeSelectedPage() {
    if (!selectedPage) throw new Error("Facebook page is unavailable");
    const result = facebookHandoffCode
      ? await connectFacebookPage(facebookHandoffCode, selectedPage.pageId)
      : await selectConnectedFacebookPage(selectedPage.pageId);
    setSelectedPage(result.page);
    setLatestReport(result.report ?? null);
    setAnalysisStatus(result.analysisStatus);
    setWeeklyReport(await getWeeklyFacebookReport(result.page.pageId));
    setHasPagePermission(true);
    setFacebookHandoffCode(null);
    void activateDashboardRichMenu().catch(() => undefined);
  }

  async function disconnectSelectedPage() {
    if (!selectedPage) return;
    await disconnectFacebookPage(selectedPage.pageId);
    void activateConnectRichMenu().catch(() => undefined);
    clearFacebookSession();
  }

  async function deleteSelectedPageData() {
    if (!selectedPage) return;
    await deleteFacebookPageData(selectedPage.pageId);
    void activateConnectRichMenu().catch(() => undefined);
    clearFacebookSession();
  }

  if (!isLineIdentityReady) {
    return (
      <MobileAppShell>
        <Box
          sx={{
            alignItems: "center",
            display: "flex",
            flexDirection: "column",
            gap: 1.5,
            justifyContent: "center",
            minHeight: "calc(100dvh - 150px)",
            textAlign: "center",
          }}
        >
          <LoadingDots />
          <Typography sx={{ fontSize: 19, fontWeight: 900 }}>กรุณารอสักครู่</Typography>
          <Typography color="text.secondary" sx={{ fontSize: 14 }}>
            กำลังตรวจสอบการเข้าใช้งานผ่าน LINE
          </Typography>
        </Box>
      </MobileAppShell>
    );
  }

  if (facebookHandoff) {
    return (
      <MobileAppShell>
        <Box
          sx={{
            alignItems: "center",
            display: "flex",
            flexDirection: "column",
            gap: 1.5,
            justifyContent: "center",
            minHeight: "calc(100dvh - 150px)",
            textAlign: "center",
          }}
          >
            <LoadingDots />
          <Typography sx={{ fontSize: 19, fontWeight: 900 }}>กรุณารอสักครู่</Typography>
          <Typography color="text.secondary" sx={{ fontSize: 14 }}>
            กำลังพาไปเลือกเพจ Facebook
          </Typography>
        </Box>
      </MobileAppShell>
    );
  }

  return (
    <MobileAppShell>
      <Box component={Paper} elevation={0} sx={{ bgcolor: "transparent" }}>
        <Suspense fallback={<RouteLoading />}>
          <Routes>
          <Route
            element={
              <Navigate replace to={canViewDashboard ? "/dashboard" : "/connect-facebook"} />
            }
            path="/"
          />
          <Route
            element={
              canViewDashboard && selectedPage ? (
                latestReport ? (
                  <DashboardPage
                    analysisStatus={analysisStatus}
                    onDeleteData={deleteSelectedPageData}
                    onDisconnect={disconnectSelectedPage}
                    page={selectedPage}
                    report={latestReport}
                    weeklyReport={weeklyReport}
                  />
                ) : (
                  <DashboardAnalysisPending page={selectedPage} status={analysisStatus} />
                )
              ) : (
                <Navigate replace to={hasFacebookLogin ? "/pages" : "/connect-facebook"} />
              )
            }
            path="/dashboard"
          />
          <Route
            element={
              <ConnectFacebookPage
                hasFacebookLogin={hasFacebookLogin}
                isLoading={isCompletingFacebookLogin}
                loginError={facebookLoginError}
                onLogin={startFacebookLogin}
              />
            }
            path="/connect-facebook"
          />
          <Route
            element={
              hasFacebookLogin ? (
                <PageSelectPage
                  onAuthorize={authorizeSelectedPage}
                  onSelectPage={(page) => {
                    setSelectedPage(page);
                    setHasPagePermission(false);
                  }}
                  selectedPage={selectedPage}
                  pages={facebookPages}
                />
              ) : (
                <Navigate replace to="/connect-facebook" />
              )
            }
            path="/pages"
          />
          <Route element={<AnalyzingPage />} path="/analyzing" />
          </Routes>
        </Suspense>
      </Box>
    </MobileAppShell>
  );
}

function PublicLegalPage({ type }: { type: "privacy" | "terms" | "data-deletion" }) {
  return (
    <MobileAppShell>
      <Suspense fallback={<RouteLoading />}>
        <LegalPage type={type} />
      </Suspense>
    </MobileAppShell>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<PublicLegalPage type="privacy" />} path="/privacy" />
        <Route element={<PublicLegalPage type="terms" />} path="/terms" />
        <Route element={<PublicLegalPage type="data-deletion" />} path="/data-deletion" />
        <Route element={<AppRoutes />} path="*" />
      </Routes>
    </BrowserRouter>
  );
}
