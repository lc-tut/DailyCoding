import { notFound } from "next/navigation";
import ProblemPanel from "@/features/problem/ui/ProblemPanel";
import SubmissionArea from "@/features/submit/ui/SubmissionArea";
import { fetchProblem } from "@/features/problem/api/fetchProblem";
import { initialCode } from "@/features/problem/model/problem";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function ProblemPage({ params }: Props) {
  const { id } = await params;

  let problem;
  try {
    problem = await fetchProblem(id);
  } catch {
    problem = null;
  }

  if (!problem) notFound();

  return (
    <>
      <ProblemPanel problem={problem} submissionsHref={`/problems/${id}/submissions`} />
      <SubmissionArea initialCode={initialCode} problemId={problem.id} />
    </>
  );
}
