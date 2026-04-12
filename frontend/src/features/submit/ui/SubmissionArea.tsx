"use client";

import dynamic from "next/dynamic";
import { useState } from "react";
import Button from "@/components/ui/Button";
import ResultPanel from "./ResultPanel";
import { submitCode } from "../api/submitCode";
import type { JudgeResponse } from "../api/submitCode";
import styles from "./SubmissionArea.module.css";

// CodeMirror は SSR 非対応のため dynamic import でクライアント専用にする
const CodeEditor = dynamic(
  () => import("@/features/editor/ui/CodeEditor"),
  { ssr: false }
);

type Props = {
  initialCode: string;
};

export default function SubmissionArea({ initialCode }: Props) {
  const [code, setCode] = useState(initialCode);
  const [response, setResponse] = useState<JudgeResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    setLoading(true);
    setError(null);
    setResponse(null);
    try {
      const data = await submitCode(code);
      setResponse(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "不明なエラー");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.wrapper}>
      {/* エディタ */}
      <section className={styles.editorSection}>
        <div className={styles.sectionHeader}>コード（Python）</div>
        <div className={styles.editorBody}>
          <CodeEditor value={code} onChange={setCode} />
        </div>
      </section>

      {/* 提出・結果エリア */}
      <footer className={styles.footer}>
        <Button onClick={handleSubmit} loading={loading}>
          提出
        </Button>

        {error && (
          <div className={styles.errorBox}>
            <strong>エラー:</strong> {error}
          </div>
        )}

        {response && <ResultPanel response={response} />}
      </footer>
    </div>
  );
}
