import Link from "next/link";
import { notFound } from "next/navigation";
import { fetchProblem } from "@/features/problem/api/fetchProblem";
import { fetchSubmissions } from "@/features/submit/api/fetchSubmissions";
import { Badge } from "@/components/ui/badge";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function SubmissionsPage({ params }: Props) {
  const { id } = await params;

  const [problem, submissions] = await Promise.all([
    fetchProblem(id).catch(() => null),
    fetchSubmissions(id).catch(() => []),
  ]);

  if (!problem) notFound();

  return (
    <div className="flex-1 overflow-y-auto p-8">
      <div className="max-w-3xl w-full mx-auto">
        {/* パンくず */}
        <nav className="text-xs text-[var(--color-subtext0)] mb-4 flex items-center gap-1.5">
          <Link href={`/problems/${id}`} className="hover:text-[var(--color-text)] transition-colors">
            {problem.title}
          </Link>
          <span>/</span>
          <span className="text-[var(--color-text)]">提出履歴</span>
        </nav>

        <h1 className="text-xl font-bold text-[var(--color-text)] mb-6">提出履歴</h1>

        {submissions.length === 0 ? (
          <p className="text-[var(--color-subtext0)]">まだ提出がありません。</p>
        ) : (
          <div className="flex flex-col gap-2">
            {submissions.map((s) => (
              <Link
                key={s.id}
                href={`/problems/${id}/submissions/${s.id}`}
                className="flex items-center justify-between px-4 py-3 rounded-lg bg-[var(--color-mantle)] border border-[var(--color-surface0)] hover:border-[var(--color-surface1)] transition-colors"
              >
                <div className="flex items-center gap-3">
                  <Badge variant={s.overallStatus === "Pass" ? "success" : "destructive"}>
                    {s.overallStatus}
                  </Badge>
                  <span className="text-sm font-mono text-[var(--color-text)]">
                    {s.score} 点
                  </span>
                  <span className="text-xs text-[var(--color-subtext0)]">
                    {s.language}
                  </span>
                </div>
                <span className="text-xs text-[var(--color-overlay0)] font-mono">
                  {new Date(s.createdAt).toLocaleString("ja-JP", { timeZone: "Asia/Tokyo" })}
                </span>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
