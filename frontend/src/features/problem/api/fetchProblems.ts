import { API_BASE } from "@/shared/config/api";
import type { Problem } from "../model/problem";

type ProblemAPIResponse = {
  id: string;
  title: string;
  description: string;
  is_active: boolean;
  published_at: string | null;
};

// fetchActiveProblems は GET /api/problems を呼び、Problem[] を返す
export async function fetchActiveProblems(): Promise<Problem[]> {
  const res = await fetch(`${API_BASE}/api/problems`, {
    cache: "no-store",
    signal: AbortSignal.timeout(3000),
  });
  if (!res.ok) {
    throw new Error(`GET /api/problems failed: ${res.status}`);
  }
  const data: ProblemAPIResponse[] = await res.json();
  return data.map((p) => ({
    id: p.id,
    title: p.title,
    description: p.description,
    isActive: p.is_active,
    publishedAt: p.published_at,
  }));
}
