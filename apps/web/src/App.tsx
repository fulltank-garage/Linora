import { useEffect, useRef, useState } from "react";
import { Box, Paper, Typography } from "@mui/material";
import { BrowserRouter, Navigate, Route, Routes, useLocation, useNavigate } from "react-router-dom";
import { MobileAppShell } from "@linora/ui";
import type { AnalysisReport, FacebookPageSummary } from "@linora/shared";
import { demoReport } from "./data/demo";
import { AnalyzingPage } from "./pages/AnalyzingPage";
import { ConnectFacebookPage } from "./pages/ConnectFacebookPage";
import { DashboardPage } from "./pages/DashboardPage";
import { LegalPage } from "./pages/LegalPage";
import { ManualAnalyzePage } from "./pages/ManualAnalyzePage";
import { PageSelectPage } from "./pages/PageSelectPage";
import {
  activateConnectRichMenu,
  activateDashboardRichMenu,
  completeFacebookLogin,
  connectFacebookPage,
  deleteFacebookPageData,
  disconnectFacebookPage,
  getConnectedFacebookPages,
  selectConnectedFacebookPage,
  startFacebookLogin,
  syncFacebookPage,
} from "./api/client";
import { LoadingDots } from "./components/LoadingDots";
import { initializeLineIdentity } from "./lib/line";

const facebookTransitionDelay = 650;

function AppRoutes() {
  const location = useLocation();
  const navigate = useNavigate();
  const facebookHandoff = new URLSearchParams(location.search).get("facebook_connect");
  const [latestReport, setLatestReport] = useState<AnalysisReport>(demoReport);
  const [hasFacebookLogin, setHasFacebookLogin] = useState(false);
  const [facebookPages, setFacebookPages] = useState<FacebookPageSummary[]>([]);
  const [facebookLoginError, setFacebookLoginError] = useState<string | null>(null);
  const [facebookHandoffCode, setFacebookHandoffCode] = useState<string | null>(null);
  const [isCompletingFacebookLogin, setIsCompletingFacebookLogin] = useState(false);
  const [isLineIdentityReady, setIsLineIdentityReady] = useState(false);
  const completedFacebookHandoff = useRef<string | null>(null);
  const [selectedPage, setSelectedPage] = useState<FacebookPageSummary | null>(null);
  const [hasPagePermission, setHasPagePermission] = useState(false);

  const canViewDashboard =
    hasFacebookLogin && selectedPage !== null && hasPagePermission;

  function clearFacebookSession() {
    setHasFacebookLogin(false);
    setFacebookPages([]);
    setFacebookHandoffCode(null);
    setSelectedPage(null);
    setHasPagePermission(false);
  }

  useEffect(() => {
    let isActive = true;
    void initializeLineIdentity()
      .then(async (ready) => {
        if (!ready || !isActive) return;
        const pages = await getConnectedFacebookPages();
        if (!isActive) return;
        setFacebookPages(pages);
        setHasFacebookLogin(pages.length > 0);
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
    setLatestReport(result.report);
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

  async function analyzeSelectedPage() {
    if (!selectedPage) throw new Error("Facebook page is unavailable");
    const report = await syncFacebookPage(selectedPage.pageId);
    setLatestReport(report);
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
        <Routes>
          <Route
            element={
              <Navigate replace to={canViewDashboard ? "/dashboard" : "/connect-facebook"} />
            }
            path="/"
          />
          <Route
            element={
              canViewDashboard ? (
                <DashboardPage
                  onAnalyze={analyzeSelectedPage}
                  onDeleteData={deleteSelectedPageData}
                  onDisconnect={disconnectSelectedPage}
                  page={selectedPage}
                  report={latestReport}
                />
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
          <Route element={<ManualAnalyzePage onReport={setLatestReport} />} path="/manual-analyze" />
          <Route element={<LegalPage type="privacy" />} path="/privacy" />
          <Route element={<LegalPage type="terms" />} path="/terms" />
          <Route element={<LegalPage type="data-deletion" />} path="/data-deletion" />
        </Routes>
      </Box>
    </MobileAppShell>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  );
}
