import axios from "axios";
import type { AnalysisReport, AnalysisStatus, FacebookPageSummary, WeeklyReport } from "@linora/shared";

export type FacebookPageResult = {
  analysisStatus: AnalysisStatus;
  page: FacebookPageSummary;
  report?: AnalysisReport;
};

export type FacebookDashboardResult = FacebookPageResult & {
  weeklyReport: WeeklyReport;
};

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? "http://localhost:8080",
});

export function setLineIdentityToken(idToken: string) {
  api.defaults.headers.common.Authorization = `Bearer ${idToken}`;
  delete api.defaults.headers.common["X-Linora-Dev-User"];
}

export function setDevelopmentLineUser(userID: string) {
  if (!import.meta.env.DEV) return;
  api.defaults.headers.common["X-Linora-Dev-User"] = userID;
}

export async function startFacebookLogin() {
  const response = await api.post<{ authorizationUrl: string }>("/api/facebook/login");
  window.location.assign(response.data.authorizationUrl);
}

export async function completeFacebookLogin(code: string) {
  const response = await api.get<{ pages: FacebookPageSummary[] }>("/api/facebook/session", {
    params: { code },
  });
  return response.data.pages;
}

export async function connectFacebookPage(handoffCode: string, pageId: string) {
  const response = await api.post<FacebookPageResult>("/api/facebook/connections", { handoffCode, pageId });
  return response.data;
}

export async function getConnectedFacebookPages() {
  const response = await api.get<{ pages: FacebookPageSummary[] }>("/api/facebook/pages");
  return response.data.pages;
}

export async function getSavedFacebookDashboard() {
  try {
    const response = await api.get<FacebookDashboardResult>("/api/facebook/dashboard");
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 404) return null;
    throw error;
  }
}

export async function getWeeklyFacebookReport(pageId: string) {
  const response = await api.get<{ weeklyReport: WeeklyReport }>(`/api/facebook/pages/${pageId}/weekly-report`);
  return response.data.weeklyReport;
}

export async function selectConnectedFacebookPage(pageId: string) {
  const response = await api.post<FacebookPageResult>(`/api/facebook/pages/${pageId}/select`);
  return response.data;
}

export async function disconnectFacebookPage(pageId: string) {
  await api.delete(`/api/facebook/pages/${pageId}/connection`);
}

export async function deleteFacebookPageData(pageId: string) {
  await api.delete(`/api/facebook/pages/${pageId}`);
}

export async function activateDashboardRichMenu() {
  await api.post("/api/line/rich-menu/dashboard");
}

export async function activateConnectRichMenu() {
  await api.post("/api/line/rich-menu/connect");
}
