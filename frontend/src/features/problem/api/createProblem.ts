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
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const text = await res.text().catch(() => String(res.status));
    throw new Error(`問題の作成に失敗しました: ${text}`);
  }
  return res.json() as Promise<CreateProblemResponse>;
}
