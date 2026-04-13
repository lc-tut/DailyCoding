import { notFound } from "next/navigation";
import { fetchProblem } from "@/features/problem/api/fetchProblem";
import ProblemEditForm from "@/features/problem/ui/ProblemEditForm";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function EditProblemPage({ params }: Props) {
  const { id } = await params;

  let problem;
  try {
    problem = await fetchProblem(id);
  } catch {
    problem = null;
  }

  if (!problem) notFound();

  return (
    <div className="flex-1 overflow-y-auto p-8">
      <h1 className="text-xl font-bold text-[var(--color-text)] mb-1">問題を編集</h1>
      <p className="text-xs text-[var(--color-overlay0)] mb-6 font-mono">{problem.id}</p>
      <ProblemEditForm problem={problem} />
    </div>
  );
}
