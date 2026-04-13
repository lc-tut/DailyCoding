import { API_BASE } from "@/shared/config/api";
import type { Problem } from "../model/problem";

type ProblemAPIResponse = {
  id: string;
  title: string;
  description: string;
  is_active: boolean;
  published_at: string | null;
};

// fetchProblem は GET /api/problems/:id を呼び、問題を1件返す
// 見つからない場合は null を返す
export async function fetchProblem(id: string): Promise<Problem | null> {
  const res = await fetch(`${API_BASE}/api/problems/${id}`, {
    next: { revalidate: 60 },
    signal: AbortSignal.timeout(3000),
  });
  if (res.status === 404) return null;
  if (!res.ok) {
    throw new Error(`GET /api/problems/${id} failed: ${res.status}`);
  }
  const p: ProblemAPIResponse = await res.json();
  return {
    id: p.id,
    title: p.title,
    description: p.description,
    isActive: p.is_active,
    publishedAt: p.published_at,
  };
}
