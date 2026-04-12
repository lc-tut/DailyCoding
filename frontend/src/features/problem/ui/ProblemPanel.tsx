import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { Problem } from "../model/problem";
import styles from "./ProblemPanel.module.css";

type Props = {
  problem: Problem;
};

// サーバーコンポーネント: 静的な問題文を SSR でレンダリングする
export default function ProblemPanel({ problem }: Props) {
  return (
    <section className={styles.panel}>
      <div className={styles.header}>問題</div>
      <div className={styles.body}>
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          components={{
            h2: ({ children }) => <h2 className={styles.h2}>{children}</h2>,
            h3: ({ children }) => <h3 className={styles.h3}>{children}</h3>,
            p: ({ children }) => <p className={styles.p}>{children}</p>,
            code: ({ children, className }) => {
              const isBlock = className?.startsWith("language-");
              return isBlock ? (
                <pre className={styles.pre}>
                  <code>{children}</code>
                </pre>
              ) : (
                <code className={styles.inlineCode}>{children}</code>
              );
            },
            table: ({ children }) => (
              <table className={styles.table}>{children}</table>
            ),
          }}
        >
          {problem.description}
        </ReactMarkdown>
      </div>
    </section>
  );
}
