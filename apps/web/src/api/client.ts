import axios from "axios";
import type { ManualAnalysisRequest, ManualAnalysisResponse } from "@linora/shared";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? "http://localhost:8080",
});

export async function runManualAnalysis(input: ManualAnalysisRequest) {
  const response = await api.post<ManualAnalysisResponse>("/api/analysis/manual", input);
  return response.data.report;
}
