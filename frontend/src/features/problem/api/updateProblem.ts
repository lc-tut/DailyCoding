import { API_BASE } from "@/shared/config/api";

export type UpdateProblemInput = {
  title: string;
  description: string;
  is_active: boolean;
  published_at?: string | null; // "YYYY-MM-DD" | null (null = 常時公開)
};

export type UpdateProblemResponse = {
  id: string;
  title: string;
  description: string;
  is_active: boolean;
  published_at: string | null;
};

export async function updateProblem(
  id: string,
  input: UpdateProblemInput
): Promise<UpdateProblemResponse> {
  const res = await fetch(`${API_BASE}/api/problems/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    const msg = body?.error ?? `ステータス ${res.status}`;
    throw new Error(msg);
  }
  return res.json() as Promise<UpdateProblemResponse>;
}
