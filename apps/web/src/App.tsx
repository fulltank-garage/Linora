import { useState } from "react";
import { Box, Paper } from "@mui/material";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { MobileAppShell } from "@linora/ui";
import type { AnalysisReport, FacebookPageSummary } from "@linora/shared";
import { demoReport } from "./data/demo";
import { AnalyzingPage } from "./pages/AnalyzingPage";
import { ConnectFacebookPage } from "./pages/ConnectFacebookPage";
import { DashboardPage } from "./pages/DashboardPage";
import { LegalPage } from "./pages/LegalPage";
import { ManualAnalyzePage } from "./pages/ManualAnalyzePage";
import { PageSelectPage } from "./pages/PageSelectPage";

function AppRoutes() {
  const [latestReport, setLatestReport] = useState<AnalysisReport>(demoReport);
  const [hasFacebookLogin, setHasFacebookLogin] = useState(false);
  const [selectedPage, setSelectedPage] = useState<FacebookPageSummary | null>(null);
  const [hasPagePermission, setHasPagePermission] = useState(false);

  const canViewDashboard =
    hasFacebookLogin && selectedPage !== null && hasPagePermission;

  function clearFacebookSession() {
    setHasFacebookLogin(false);
    setSelectedPage(null);
    setHasPagePermission(false);
  }

  return (
    <MobileAppShell>
      <Box component={Paper} elevation={0} sx={{ bgcolor: "transparent" }}>
        <Routes>
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
            path="/"
          />
          <Route
            element={
              <ConnectFacebookPage
                hasFacebookLogin={hasFacebookLogin}
                onLogin={() => setHasFacebookLogin(true)}
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
