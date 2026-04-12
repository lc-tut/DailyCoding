"use client";

import dynamic from "next/dynamic";
import { useState } from "react";
import styles from "./ProblemPage.module.css";

// CodeMirrorはSSRと相性が悪いためdynamic importでクライアント専用にする
const CodeEditor = dynamic(() => import("./CodeEditor"), { ssr: false });

type TestResult = {
  case_index: number;
  status: "Pass" | "Fail";
  stdout: string;
  expected: string;
  stderr: string;
  time_ms: number;
};

type JudgeResponse = {
  overall: "Pass" | "Fail";
  results: TestResult[];
};

const INITIAL_CODE = `a, b = map(int, input().split())
print(a + b)
`;

const PROBLEM_STATEMENT = `## 問題：二数の和

標準入力から空白区切りで2つの整数を受け取り、
その和を標準出力に出力してください。

### 入力形式
\`\`\`
A B
\`\`\`

### 出力形式
\`\`\`
A + B の値
\`\`\`

### サンプル入力1
\`\`\`
2 3
\`\`\`
### サンプル出力1
\`\`\`
5
\`\`\`

### サンプル入力2
\`\`\`
10 20
\`\`\`
### サンプル出力2
\`\`\`
30
\`\`\`
`;

export default function ProblemPage() {
  const [code, setCode] = useState(INITIAL_CODE);
  const [response, setResponse] = useState<JudgeResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    setLoading(true);
    setError(null);
    setResponse(null);
    try {
      const res = await fetch("http://localhost:8080/api/submit", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code }),
      });
      if (!res.ok) {
        throw new Error(`サーバーエラー: ${res.status}`);
      }
      const data: JudgeResponse = await res.json();
      setResponse(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "不明なエラー");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.container}>
      {/* ヘッダー */}
      <header className={styles.header}>
        <span className={styles.logo}>DaileyCoding</span>
        <span className={styles.subtitle}>LinuxClub Python Challenge</span>
      </header>

      {/* メインエリア */}
      <main className={styles.main}>
        {/* 左: 問題文 */}
        <section className={styles.problem}>
          <h2 className={styles.sectionTitle}>問題</h2>
          <div className={styles.problemContent}>
            {PROBLEM_STATEMENT.split("\n").map((line, i) => {
              if (line.startsWith("## ")) {
                return <h3 key={i} className={styles.problemH3}>{line.slice(3)}</h3>;
              }
              if (line.startsWith("### ")) {
                return <h4 key={i} className={styles.problemH4}>{line.slice(4)}</h4>;
              }
              if (line.startsWith("```")) {
                return null;
              }
              return <p key={i} className={styles.problemText}>{line}</p>;
            })}
          </div>
        </section>

        {/* 右: エディタ */}
        <section className={styles.editor}>
          <h2 className={styles.sectionTitle}>コード（Python）</h2>
          <div className={styles.editorWrapper}>
            <CodeEditor value={code} onChange={setCode} />
          </div>
        </section>
      </main>

      {/* フッター: 提出ボタン + 結果 */}
      <footer className={styles.footer}>
        <button
          className={styles.submitButton}
          onClick={handleSubmit}
          disabled={loading}
        >
          {loading ? "実行中..." : "提出"}
        </button>

        {error && (
          <div className={styles.errorBox}>
            <strong>エラー:</strong> {error}
          </div>
        )}

        {response && (
          <div className={styles.resultBox}>
            <span
              className={
                response.overall === "Pass"
                  ? styles.overallPass
                  : styles.overallFail
              }
            >
              {response.overall === "Pass" ? "✓ 正解" : "✗ 不正解"}
            </span>
            <div className={styles.cases}>
              {response.results.map((r) => (
                <div
                  key={r.case_index}
                  className={
                    r.status === "Pass" ? styles.casePass : styles.caseFail
                  }
                >
                  <span className={styles.caseLabel}>
                    Case {r.case_index}: {r.status}
                  </span>
                  <span className={styles.caseTime}>
                    {r.time_ms.toFixed(0)}ms
                  </span>
                  {r.status === "Fail" && (
                    <div className={styles.caseDiff}>
                      {r.stderr ? (
                        <span>stderr: {r.stderr}</span>
                      ) : (
                        <>
                          <span>出力: {JSON.stringify(r.stdout.trim())}</span>
                          <span>期待: {JSON.stringify(r.expected.trim())}</span>
                        </>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </footer>
    </div>
  );
}
