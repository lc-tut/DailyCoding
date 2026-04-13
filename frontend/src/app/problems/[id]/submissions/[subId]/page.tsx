import Link from "next/link";
import { notFound } from "next/navigation";
import { fetchSubmission } from "@/features/submit/api/fetchSubmission";
import { Badge } from "@/components/ui/badge";

type Props = {
  params: Promise<{ id: string; subId: string }>;
};

export default async function SubmissionDetailPage({ params }: Props) {
  const { id, subId } = await params;

  const submission = await fetchSubmission(subId).catch(() => null);
  if (!submission) notFound();

  return (
    <div className="flex-1 overflow-y-auto p-8">
      <div className="max-w-3xl w-full mx-auto flex flex-col gap-6">
        {/* パンくず */}
        <nav className="text-xs text-[var(--color-subtext0)] flex items-center gap-1.5">
          <Link href={`/problems/${id}`} className="hover:text-[var(--color-text)] transition-colors">
            問題
          </Link>
          <span>/</span>
          <Link href={`/problems/${id}/submissions`} className="hover:text-[var(--color-text)] transition-colors">
            提出履歴
          </Link>
          <span>/</span>
          <span className="text-[var(--color-text)] font-mono">{submission.id.slice(0, 8)}…</span>
        </nav>

        {/* サマリ */}
        <div className="flex items-center gap-4 p-4 rounded-lg bg-[var(--color-mantle)] border border-[var(--color-surface0)]">
          <Badge variant={submission.overallStatus === "Pass" ? "success" : "destructive"} className="text-sm px-3 py-1">
            {submission.overallStatus}
          </Badge>
          <span className="text-2xl font-bold text-[var(--color-text)]">{submission.score} 点</span>
          <span className="text-xs text-[var(--color-overlay0)] font-mono ml-auto">
            {new Date(submission.createdAt).toLocaleString("ja-JP", { timeZone: "Asia/Tokyo" })}
          </span>
        </div>

        {/* 提出コード */}
        <section className="flex flex-col gap-2">
          <h2 className="text-sm font-semibold text-[var(--color-text)]">提出コード</h2>
          <pre className="p-4 rounded-lg bg-[var(--color-crust)] border border-[var(--color-surface0)] text-xs font-mono text-[var(--color-text)] overflow-x-auto whitespace-pre">
            {submission.code}
          </pre>
        </section>

        {/* テストケース結果 */}
        <section className="flex flex-col gap-2">
          <h2 className="text-sm font-semibold text-[var(--color-text)]">テストケース結果</h2>
          <div className="flex flex-col gap-2">
            {submission.results.map((r) => (
              <div
                key={r.orderIndex}
                className="rounded-lg border border-[var(--color-surface0)] bg-[var(--color-mantle)] overflow-hidden"
              >
                {/* ヘッダー */}
                <div className="flex items-center gap-3 px-4 py-2 border-b border-[var(--color-surface0)]">
                  <Badge variant={r.status === "Pass" ? "success" : "destructive"}>
                    {r.status}
                  </Badge>
                  <span className="text-xs text-[var(--color-subtext0)]">
                    ケース {r.orderIndex}
                    {r.isSample && <span className="ml-1 text-[var(--color-overlay0)]">(サンプル)</span>}
                  </span>
                  <span className="text-xs text-[var(--color-overlay0)] ml-auto font-mono">
                    {r.executionTimeMs.toFixed(1)} ms
                  </span>
                </div>

                {/* 入出力（サンプルのみ） */}
                <div className="grid grid-cols-1 divide-y divide-[var(--color-surface0)]">
                  {r.isSample && (
                    <>
                      <Row label="入力" value={r.stdin} />
                      <Row label="期待値" value={r.expectedStdout} />
                    </>
                  )}
                  <Row label="出力" value={r.stdout || "(なし)"} />
                  {r.stderr && <Row label="標準エラー" value={r.stderr} highlight />}
                </div>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  );
}

function Row({
  label,
  value,
  highlight = false,
}: {
  label: string;
  value: string;
  highlight?: boolean;
}) {
  return (
    <div className="px-4 py-2 flex gap-3">
      <span className={`text-xs w-20 shrink-0 font-medium ${highlight ? "text-[var(--color-red)]" : "text-[var(--color-subtext0)]"}`}>
        {label}
      </span>
      <pre className={`text-xs font-mono whitespace-pre-wrap break-all ${highlight ? "text-[var(--color-red)]" : "text-[var(--color-text)]"}`}>
        {value}
      </pre>
    </div>
  );
}
