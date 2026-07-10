import axios from "axios";
import type { FacebookPageSummary, ManualAnalysisRequest, ManualAnalysisResponse } from "@linora/shared";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? "http://localhost:8080",
});

export async function runManualAnalysis(input: ManualAnalysisRequest) {
  const response = await api.post<ManualAnalysisResponse>("/api/analysis/manual", input);
  return response.data.report;
}

export function startFacebookLogin() {
  window.location.assign(`${api.defaults.baseURL}/api/facebook/login`);
}

export async function completeFacebookLogin(code: string) {
  const response = await api.get<{ pages: FacebookPageSummary[] }>("/api/facebook/session", {
    params: { code },
  });
  return response.data.pages;
}
