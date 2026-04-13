import { API_BASE } from "@/shared/config/api";
import type { SubmissionSummary } from "../model/submission";

type SubmissionSummaryAPIResponse = {
  id: string;
  language: string;
  overall_status: string;
  score: number;
  created_at: string;
};

export async function fetchSubmissions(problemId: string): Promise<SubmissionSummary[]> {
  const res = await fetch(`${API_BASE}/api/problems/${problemId}/submissions`, {
    cache: "no-store",
    signal: AbortSignal.timeout(3000),
  });
  if (!res.ok) {
    throw new Error(`GET /api/problems/${problemId}/submissions failed: ${res.status}`);
  }
  const data: SubmissionSummaryAPIResponse[] = await res.json();
  return data.map((s) => ({
    id: s.id,
    language: s.language,
    overallStatus: s.overall_status as "Pass" | "Fail",
    score: s.score,
    createdAt: s.created_at,
  }));
}
