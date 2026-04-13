import Link from "next/link";
import type { Problem } from "@/features/problem/model/problem";
import { fetchActiveProblems } from "@/features/problem/api/fetchProblems";
import { Badge } from "@/components/ui/badge";

export default async function ProblemsPage() {
  let problems: Problem[];
  try {
    problems = await fetchActiveProblems();
  } catch {
    problems = [];
  }

  return (
    <div className="flex-1 overflow-y-auto p-8">
      <h1 className="text-xl font-bold text-[var(--color-text)] mb-6">問題一覧</h1>

      {problems.length === 0 ? (
        <p className="text-[var(--color-subtext0)]">公開中の問題はまだありません。</p>
      ) : (
        <div className="flex flex-col gap-2 max-w-2xl">
          {problems.map((p) => (
            <Link
              key={p.id}
              href={`/problems/${p.id}`}
              className="flex items-center justify-between px-4 py-3 rounded-lg bg-[var(--color-mantle)] border border-[var(--color-surface0)] hover:border-[var(--color-mauve)] transition-colors group"
            >
              <div className="flex items-center gap-3">
                {p.publishedAt && (
                  <span className="text-xs text-[var(--color-overlay0)] font-mono w-24 shrink-0">
                    {p.publishedAt}
                  </span>
                )}
                <span className="text-sm font-medium text-[var(--color-text)] group-hover:text-[var(--color-mauve)] transition-colors">
                  {p.title}
                </span>
              </div>
              {!p.publishedAt && (
                <Badge variant="secondary" className="text-xs">常時公開</Badge>
              )}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
