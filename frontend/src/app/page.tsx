import Link from "next/link";
import ProblemPanel from "@/features/problem/ui/ProblemPanel";
import SubmissionArea from "@/features/submit/ui/SubmissionArea";
import { fetchTodayProblem } from "@/features/problem/api/fetchTodayProblem";
import { initialCode } from "@/features/problem/model/problem";

export default async function Home() {
  let problem;
  try {
    problem = await fetchTodayProblem();
  } catch {
    problem = null;
  }

  if (problem === null) {
    // 今日の問題がない場合はメッセージを表示して一覧へ誘導
    // redirect() の代わりにインラインで表示（リダイレクトは UX が荒いため）
    return (
      <div className="flex flex-col items-center justify-center flex-1 gap-4 text-center p-8">
        <p className="text-[var(--color-subtext0)] text-base">
          今日の問題はまだ公開されていません。
        </p>
        <Link
          href="/problems"
          className="text-sm text-[var(--color-mauve)] hover:text-[var(--color-lavender)] underline underline-offset-4 transition-colors"
        >
          過去の問題一覧を見る →
        </Link>
      </div>
    );
  }

  return (
    <>
      <ProblemPanel problem={problem} />
      <SubmissionArea initialCode={initialCode} problemId={problem.id} />
    </>
  );
}
