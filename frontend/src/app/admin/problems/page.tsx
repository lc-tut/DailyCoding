import Link from "next/link";
import type { Problem } from "@/features/problem/model/problem";
import { fetchAllProblems } from "@/features/problem/api/fetchAllProblems";
import { Button } from "@/components/ui/button";
import { AdminProblemsClient } from "./AdminProblemsClient";

type SearchParams = Promise<{ filter?: string; sort?: string }>;

export default async function AdminProblemsPage({
  searchParams,
}: {
  searchParams: SearchParams;
}) {
  const { filter, sort } = await searchParams;
  let problems: Problem[];
  try {
    problems = await fetchAllProblems({ filter, sort });
  } catch {
    problems = [];
  }

  return (
    <div className="flex-1 overflow-y-auto p-8">
      <div className="flex items-center justify-between mb-6 max-w-3xl">
        <h1 className="text-xl font-bold text-[var(--color-text)]">問題管理</h1>
        <Button asChild size="sm">
          <Link href="/admin/problems/new">+ 新規作成</Link>
        </Button>
      </div>

      <div className="max-w-3xl">
        <AdminProblemsClient
          problems={problems}
          currentFilter={filter ?? "all"}
          currentSort={sort ?? "created"}
        />
      </div>
    </div>
  );
}
