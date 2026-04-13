import { API_BASE } from "@/shared/config/api";
import type { Problem } from "../model/problem";

type ProblemAPIResponse = {
  id: string;
  title: string;
  description: string;
  is_active: boolean;
  published_at: string | null;
};

// fetchTodayProblem は GET /api/problems/today を呼び、今日の問題を返す
// 今日の問題がない場合は null を返す
export async function fetchTodayProblem(): Promise<Problem | null> {
  const res = await fetch(`${API_BASE}/api/problems/today`, {
    // 1分ごとに再検証（日付が変わっても反映される）
    next: { revalidate: 60 },
    signal: AbortSignal.timeout(3000),
  });
  if (res.status === 404) return null;
  if (!res.ok) {
    throw new Error(`GET /api/problems/today failed: ${res.status}`);
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
