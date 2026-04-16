import { API_BASE } from "@/shared/config/api";

export type TestResult = {
  case_index: number;
  status: "Pass" | "Fail";
  stdout: string;
  expected: string;
  stderr: string;
  time_ms: number;
};

export type JudgeResponse = {
  overall: "Pass" | "Fail";
  results: TestResult[];
};

export class UnauthorizedError extends Error {
  constructor() {
    super("unauthorized");
  }
}

export async function submitCode(code: string, problemId: string): Promise<JudgeResponse> {
  const res = await fetch(`${API_BASE}/api/submit`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ code, problem_id: problemId }),
  });
  if (res.status === 401) throw new UnauthorizedError();
  if (!res.ok) {
    const text = await res.text().catch(() => String(res.status));
    throw new Error(`サーバーエラー: ${text}`);
  }
  return res.json() as Promise<JudgeResponse>;
}
