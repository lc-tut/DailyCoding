"use client";

import Link from "next/link";
import { useRouter, usePathname } from "next/navigation";
import type { Problem } from "@/features/problem/model/problem";
import { isEffectivelyActive } from "@/features/problem/model/problem";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

type FilterKey = "all" | "active" | "inactive" | "undated" | "inactive_undated";
type SortKey = "created" | "published" | "title";

const FILTER_LABELS: Record<FilterKey, string> = {
  all: "すべて",
  active: "公開中",
  inactive: "非公開",
  undated: "未日付指定",
  inactive_undated: "非公開・未日付",
};

const SORT_LABELS: Record<SortKey, string> = {
  created: "作成日",
  published: "公開日",
  title: "タイトル",
};

type Props = {
  problems: Problem[];
  currentFilter: string;
  currentSort: string;
};

export function AdminProblemsClient({ problems, currentFilter, currentSort }: Props) {
  const router = useRouter();
  const pathname = usePathname();

  const navigate = (filter: string, sort: string) => {
    const params = new URLSearchParams();
    if (filter !== "all") params.set("filter", filter);
    if (sort !== "created") params.set("sort", sort);
    const qs = params.toString();
    router.push(qs ? `${pathname}?${qs}` : pathname);
  };

  return (
    <div>
      {/* フィルタ + ソート バー */}
      <div className="flex flex-wrap items-center gap-4 mb-4">
        <div className="flex items-center gap-1">
          <span className="text-xs text-[var(--color-subtext0)] mr-1">絞込:</span>
          {(Object.keys(FILTER_LABELS) as FilterKey[]).map((key) => (
            <button
              key={key}
              onClick={() => navigate(key, currentSort)}
              className={[
                "px-3 py-1 rounded-full text-xs font-medium transition-colors",
                currentFilter === key
                  ? "bg-[var(--color-mauve)] text-[var(--color-base)]"
                  : "bg-[var(--color-surface0)] text-[var(--color-subtext1)] hover:bg-[var(--color-surface1)]",
              ].join(" ")}
            >
              {FILTER_LABELS[key]}
            </button>
          ))}
        </div>

        <div className="flex items-center gap-1">
          <span className="text-xs text-[var(--color-subtext0)] mr-1">並替:</span>
          {(Object.keys(SORT_LABELS) as SortKey[]).map((key) => (
            <button
              key={key}
              onClick={() => navigate(currentFilter, key)}
              className={[
                "px-3 py-1 rounded-full text-xs font-medium transition-colors",
                currentSort === key
                  ? "bg-[var(--color-blue)] text-[var(--color-base)]"
                  : "bg-[var(--color-surface0)] text-[var(--color-subtext1)] hover:bg-[var(--color-surface1)]",
              ].join(" ")}
            >
              {SORT_LABELS[key]}
            </button>
          ))}
        </div>

        <span className="text-xs text-[var(--color-overlay0)] ml-auto">
          {problems.length} 件
        </span>
      </div>

      {/* 問題リスト */}
      {problems.length === 0 ? (
        <p className="text-[var(--color-subtext0)]">該当する問題がありません。</p>
      ) : (
        <div className="flex flex-col gap-2">
          {problems.map((p) => {
            const active = isEffectivelyActive(p);
            return (
              <div
                key={p.id}
                className="flex items-center justify-between px-4 py-3 rounded-lg bg-[var(--color-mantle)] border border-[var(--color-surface0)]"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <span className="text-xs text-[var(--color-overlay0)] font-mono w-24 shrink-0">
                    {p.publishedAt ?? "—"}
                  </span>
                  <Link
                    href={`/problems/${p.id}`}
                    className="text-sm font-medium text-[var(--color-text)] hover:text-[var(--color-mauve)] truncate transition-colors"
                  >
                    {p.title}
                  </Link>
                </div>
                <div className="flex items-center gap-2 shrink-0 ml-4">
                  {p.publishedAt && <Badge variant="outline">日付指定</Badge>}
                  {active ? (
                    <Badge variant="default">公開中</Badge>
                  ) : (
                    <Badge variant="secondary">非公開</Badge>
                  )}
                  <Button asChild size="sm" variant="ghost">
                    <Link href={`/admin/problems/${p.id}/edit`}>編集</Link>
                  </Button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
