"use client";

import dynamic from "next/dynamic";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import ResultPanel from "./ResultPanel";
import { submitCode, UnauthorizedError } from "../api/submitCode";
import type { JudgeResponse } from "../api/submitCode";

const CodeEditor = dynamic(
  () => import("@/features/editor/ui/CodeEditor"),
  { ssr: false }
);

type Props = {
  initialCode: string;
  problemId: string;
};

export default function SubmissionArea({ initialCode, problemId }: Props) {
  const router = useRouter();
  const draftKey = `draft_code_${problemId}`;

  const [code, setCode] = useState(initialCode);
  const [response, setResponse] = useState<JudgeResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // ログイン後に復帰したとき sessionStorage からコードを復元する
  useEffect(() => {
    const saved = sessionStorage.getItem(draftKey);
    if (saved !== null) {
      setCode(saved);
      sessionStorage.removeItem(draftKey);
    }
  }, [draftKey]);

  const handleSubmit = async () => {
    setLoading(true);
    setError(null);
    setResponse(null);
    try {
      const data = await submitCode(code, problemId);
      setResponse(data);
    } catch (e) {
      if (e instanceof UnauthorizedError) {
        sessionStorage.setItem(draftKey, code);
        router.push(`/login?next=${encodeURIComponent(window.location.pathname)}`);
        return;
      }
      setError(e instanceof Error ? e.message : "不明なエラー");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* エディタ */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="text-xs font-semibold uppercase tracking-widest text-[var(--color-overlay0)] px-4 py-2 bg-[var(--color-mantle)] border-b border-[var(--color-surface0)] shrink-0">
          コード（Python）
        </div>
        <div className="flex-1 overflow-hidden">
          <CodeEditor value={code} onChange={setCode} />
        </div>
      </div>

      {/* 提出・結果エリア */}
      <footer className="flex items-start gap-4 flex-wrap px-4 py-3 bg-[var(--color-mantle)] border-t border-[var(--color-surface0)] shrink-0">
        <Button onClick={handleSubmit} disabled={loading}>
          {loading ? "実行中..." : "提出"}
        </Button>

        {error && (
          <div className="text-xs text-[var(--color-red)] px-3 py-1.5 bg-[color-mix(in_srgb,var(--color-red)_12%,transparent)] border border-[var(--color-red)] rounded">
            <strong>エラー:</strong> {error}
          </div>
        )}

        {response && <ResultPanel response={response} />}
      </footer>
    </div>
  );
}
