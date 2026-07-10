import { useEffect, useState } from "react";
import { Box, Paper } from "@mui/material";
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
import { completeFacebookLogin, startFacebookLogin } from "./api/client";

function AppRoutes() {
  const location = useLocation();
  const navigate = useNavigate();
  const [latestReport, setLatestReport] = useState<AnalysisReport>(demoReport);
  const [hasFacebookLogin, setHasFacebookLogin] = useState(false);
  const [facebookPages, setFacebookPages] = useState<FacebookPageSummary[]>([]);
  const [facebookLoginError, setFacebookLoginError] = useState<string | null>(null);
  const [isCompletingFacebookLogin, setIsCompletingFacebookLogin] = useState(false);
  const [selectedPage, setSelectedPage] = useState<FacebookPageSummary | null>(null);
  const [hasPagePermission, setHasPagePermission] = useState(false);

  const canViewDashboard =
    hasFacebookLogin && selectedPage !== null && hasPagePermission;

  function clearFacebookSession() {
    setHasFacebookLogin(false);
    setFacebookPages([]);
    setSelectedPage(null);
    setHasPagePermission(false);
  }

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const loginError = params.get("facebook_error");
    const handoffCode = params.get("facebook_connect");
    if (loginError) {
      setFacebookLoginError(loginError);
      navigate("/connect-facebook", { replace: true });
      return;
    }
    if (!handoffCode) return;

    setIsCompletingFacebookLogin(true);
    setFacebookLoginError(null);
    void completeFacebookLogin(handoffCode)
      .then((pages) => {
        if (pages.length === 0) {
          setFacebookLoginError("ไม่พบ Facebook Page ที่คุณมีสิทธิ์จัดการ");
          navigate("/connect-facebook", { replace: true });
          return;
        }
        setFacebookPages(pages);
        setHasFacebookLogin(true);
        navigate("/pages", { replace: true });
      })
      .catch(() => {
        setFacebookLoginError("ไม่สามารถรับข้อมูล Facebook Page ได้ กรุณาลองใหม่อีกครั้ง");
        navigate("/connect-facebook", { replace: true });
      })
      .finally(() => setIsCompletingFacebookLogin(false));
  }, [location.search, navigate]);

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
                  onDeleteData={clearFacebookSession}
                  onDisconnect={clearFacebookSession}
                  page={selectedPage}
                  report={latestReport}
                />
              ) : (
                <Navigate replace to="/connect-facebook" />
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
                  onAuthorize={() => setHasPagePermission(true)}
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
