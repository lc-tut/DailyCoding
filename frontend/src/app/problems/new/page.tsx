import ProblemForm from "@/features/problem/ui/ProblemForm";

export default function NewProblemPage() {
  return (
    <div className="flex-1 overflow-y-auto p-8">
      <h1 className="text-xl font-bold text-[var(--color-text)] mb-6">問題を投稿する</h1>
      <ProblemForm />
    </div>
  );
}
