import { API_BASE } from "@/shared/config/api";

export type CreateTestCaseInput = {
  stdin: string;
  expected_stdout: string;
  order_index: number;
  is_sample: boolean;
  score_weight: number;
  hint?: string;
};

export type CreateProblemInput = {
  title: string;
  description: string;
  is_active: boolean;
  published_at?: string; // "YYYY-MM-DD"
  test_cases: CreateTestCaseInput[];
};

export type CreateProblemResponse = {
  id: string;
  title: string;
  description: string;
  is_active: boolean;
  published_at: string | null;
};

export async function createProblem(
  input: CreateProblemInput
): Promise<CreateProblemResponse> {
  const res = await fetch(`${API_BASE}/api/problems`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    const msg = body?.error ?? `ステータス ${res.status}`;
    throw new Error(msg);
  }
  return res.json() as Promise<CreateProblemResponse>;
}
