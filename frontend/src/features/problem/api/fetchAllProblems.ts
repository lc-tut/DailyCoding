import { cookies } from "next/headers";
import { API_BASE } from "@/shared/config/api";
import type { Problem } from "../model/problem";

type ProblemAPIResponse = {
  id: string;
  title: string;
  description: string;
  is_active: boolean;
  published_at: string | null;
};

export type FetchAllProblemsParams = {
  filter?: string;
  sort?: string;
};

// fetchAllProblems は GET /api/admin/problems を呼ぶ（サーバーコンポーネント専用）。
// filter / sort はバックエンドで処理される。
export async function fetchAllProblems(params?: FetchAllProblemsParams): Promise<Problem[]> {
  const url = new URL(`${API_BASE}/api/admin/problems`);
  if (params?.filter && params.filter !== "all") {
    url.searchParams.set("filter", params.filter);
  }
  if (params?.sort && params.sort !== "created") {
    url.searchParams.set("sort", params.sort);
  }

  const cookieStore = await cookies();

  const res = await fetch(url.toString(), {
    cache: "no-store",
    signal: AbortSignal.timeout(3000),
    headers: {
      Cookie: cookieStore.toString(),
    },
  });
  if (!res.ok) {
    throw new Error(`GET /api/admin/problems failed: ${res.status}`);
  }
  const data: ProblemAPIResponse[] = await res.json();
  return data.map((p) => ({
    id: p.id,
    title: p.title,
    description: p.description,
    isActive: p.is_active,
    publishedAt: p.published_at ?? null,
  }));
}
