import { API_BASE } from "@/shared/config/api";
import type { SubmissionDetail, TestCaseResult } from "../model/submission";

type TestCaseResultAPIResponse = {
  order_index: number;
  is_sample: boolean;
  stdin: string;
  expected_stdout: string;
  status: string;
  stdout: string;
  stderr: string;
  execution_time_ms: number;
};

type SubmissionDetailAPIResponse = {
  id: string;
  problem_id: string;
  code: string;
  language: string;
  overall_status: string;
  score: number;
  created_at: string;
  results: TestCaseResultAPIResponse[];
};

// null を返す: 404 (存在しない提出ID)
export async function fetchSubmission(id: string): Promise<SubmissionDetail | null> {
  const res = await fetch(`${API_BASE}/api/submissions/${id}`, {
    cache: "no-store",
    signal: AbortSignal.timeout(3000),
  });
  if (res.status === 404) return null;
  if (!res.ok) {
    throw new Error(`GET /api/submissions/${id} failed: ${res.status}`);
  }
  const data: SubmissionDetailAPIResponse = await res.json();
  return {
    id: data.id,
    problemId: data.problem_id,
    code: data.code,
    language: data.language,
    overallStatus: data.overall_status as "Pass" | "Fail",
    score: data.score,
    createdAt: data.created_at,
    results: data.results.map((r): TestCaseResult => ({
      orderIndex: r.order_index,
      isSample: r.is_sample,
      stdin: r.stdin ?? "",
      expectedStdout: r.expected_stdout ?? "",
      status: r.status as "Pass" | "Fail",
      stdout: r.stdout,
      stderr: r.stderr,
      executionTimeMs: r.execution_time_ms,
    })),
  };
}
