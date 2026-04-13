import Link from "next/link";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { Problem } from "../model/problem";

type Props = {
  problem: Problem;
  /** 提出履歴ページへのパス。省略時はリンクを表示しない */
  submissionsHref?: string;
};

export default function ProblemPanel({ problem, submissionsHref }: Props) {
  return (
    <section className="flex flex-col w-[40%] border-r border-[var(--color-surface0)] overflow-y-auto shrink-0">
      <div className="flex items-center justify-between text-xs font-semibold uppercase tracking-widest text-[var(--color-overlay0)] px-4 py-2 bg-[var(--color-mantle)] border-b border-[var(--color-surface0)]">
        <span>問題</span>
        {submissionsHref && (
          <Link
            href={submissionsHref}
            className="normal-case tracking-normal text-[var(--color-subtext0)] hover:text-[var(--color-mauve)] transition-colors"
          >
            提出履歴 →
          </Link>
        )}
      </div>
      <div className="p-4 leading-7 text-sm text-[var(--color-text)]">
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          components={{
            h2: ({ children }) => (
              <h2 className="text-base font-bold text-[var(--color-mauve)] mb-3">{children}</h2>
            ),
            h3: ({ children }) => (
              <h3 className="text-sm font-semibold text-[var(--color-blue)] mt-4 mb-1.5">{children}</h3>
            ),
            p: ({ children }) => (
              <p className="mb-2 text-[var(--color-text)]">{children}</p>
            ),
            code: ({ children, className }) => {
              const isBlock = className?.startsWith("language-");
              return isBlock ? (
                <pre className="bg-[var(--color-mantle)] border border-[var(--color-surface0)] rounded px-3 py-2 my-1.5 font-mono text-xs text-[var(--color-green)] overflow-x-auto">
                  <code>{children}</code>
                </pre>
              ) : (
                <code className="bg-[var(--color-surface0)] rounded px-1 py-px font-mono text-xs text-[var(--color-red)]">
                  {children}
                </code>
              );
            },
            table: ({ children }) => (
              <table className="border-collapse text-xs my-2 w-auto">{children}</table>
            ),
            th: ({ children }) => (
              <th className="border border-[var(--color-surface0)] bg-[var(--color-mantle)] px-3 py-1 text-[var(--color-blue)]">
                {children}
              </th>
            ),
            td: ({ children }) => (
              <td className="border border-[var(--color-surface0)] px-3 py-1 text-[var(--color-text)]">
                {children}
              </td>
            ),
          }}
        >
          {problem.description}
        </ReactMarkdown>
      </div>
    </section>
  );
}
