"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { useUser } from "@/shared/providers/UserProvider";
import {
  createProblem,
  type CreateTestCaseInput,
} from "../api/createProblem";

type PublishMode = "dated" | "open";

type TestCaseField = Omit<CreateTestCaseInput, "order_index">;

const emptyTestCase = (): TestCaseField => ({
  stdin: "",
  expected_stdout: "",
  is_sample: false,
  score_weight: 1,
});

export default function ProblemForm() {
  const router = useRouter();
  const user = useUser();
  const isAdmin = user.role === "admin";

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [isActive, setIsActive] = useState(false);
  const [publishMode, setPublishMode] = useState<PublishMode>("dated");
  const [publishedAt, setPublishedAt] = useState("");
  const [testCases, setTestCases] = useState<TestCaseField[]>([emptyTestCase()]);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const updateTestCase = (index: number, patch: Partial<TestCaseField>) =>
    setTestCases((prev) =>
      prev.map((tc, i) => (i === index ? { ...tc, ...patch } : tc))
    );

  const handleSubmit: React.EventHandler<React.SyntheticEvent<HTMLFormElement>> = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await createProblem({
        title,
        description,
        // admin のみ公開設定を送る。student は常に下書き（is_active: false）
        is_active: isAdmin ? isActive : false,
        published_at:
          isAdmin && publishMode === "dated" && publishedAt
            ? publishedAt
            : undefined,
        test_cases: testCases.map((tc, i) => ({ ...tc, order_index: i + 1 })),
      });
      // 投稿成功後、管理者は管理一覧へ・一般は問題一覧へ遷移
      router.push(isAdmin ? "/admin/problems" : "/problems");
      router.refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "不明なエラー");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6 max-w-2xl w-full">

      {/* タイトル */}
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="title">タイトル</Label>
        <Input
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
          placeholder="例: 二数の和"
        />
      </div>

      {/* 問題文 */}
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="description">問題文（Markdown）</Label>
        <Textarea
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={12}
          className="font-mono text-xs"
          placeholder={"## 問題\n標準入力から..."}
        />
      </div>

      {/* 公開設定 — admin のみ表示 */}
      {isAdmin && (
        <fieldset className="flex flex-col gap-3 p-4 rounded-lg border border-[var(--color-surface0)] bg-[var(--color-mantle)]">
          <legend className="text-sm font-semibold text-[var(--color-text)] px-1">公開設定</legend>

          <div className="flex gap-6 flex-wrap">
            <label className="flex items-center gap-2 cursor-pointer text-sm text-[var(--color-text)]">
              <input
                type="radio"
                name="publishMode"
                value="dated"
                checked={publishMode === "dated"}
                onChange={() => setPublishMode("dated")}
                className="accent-[var(--color-mauve)]"
              />
              日付指定（今日の問題として表示）
            </label>
            <label className="flex items-center gap-2 cursor-pointer text-sm text-[var(--color-text)]">
              <input
                type="radio"
                name="publishMode"
                value="open"
                checked={publishMode === "open"}
                onChange={() => setPublishMode("open")}
                className="accent-[var(--color-mauve)]"
              />
              常時公開（日付なし）
            </label>
          </div>

          {publishMode === "dated" && (
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="publishedAt">公開日</Label>
              <Input
                id="publishedAt"
                type="date"
                value={publishedAt}
                onChange={(e) => setPublishedAt(e.target.value)}
                className="w-44 cursor-pointer"
              />
            </div>
          )}

          <div className="flex items-center gap-2">
            <Checkbox
              id="isActive"
              checked={isActive}
              onCheckedChange={(v) => setIsActive(v === true)}
            />
            <Label htmlFor="isActive" className="cursor-pointer font-normal">
              即時公開する（is_active = true）
            </Label>
          </div>
        </fieldset>
      )}

      {!isAdmin && (
        <p className="text-xs text-[var(--color-overlay0)] bg-[var(--color-mantle)] border border-[var(--color-surface0)] rounded px-3 py-2">
          投稿した問題は管理者が確認後に公開されます。
        </p>
      )}

      {/* テストケース */}
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <Label>テストケース</Label>
          <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={() => setTestCases((p) => [...p, emptyTestCase()])}
          >
            + 追加
          </Button>
        </div>

        {testCases.map((tc, i) => (
          <div
            key={i}
            className="flex flex-col gap-3 p-3 rounded-lg border border-[var(--color-surface0)] bg-[var(--color-mantle)]"
          >
            <div className="flex items-center justify-between text-xs font-semibold text-[var(--color-subtext0)]">
              <span>ケース {i + 1}</span>
              {testCases.length > 1 && (
                <button
                  type="button"
                  onClick={() => setTestCases((p) => p.filter((_, j) => j !== i))}
                  className="text-[var(--color-red)] hover:underline text-xs"
                >
                  削除
                </button>
              )}
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-[var(--color-subtext0)]">標準入力 (stdin)</Label>
                <Textarea
                  value={tc.stdin}
                  onChange={(e) => updateTestCase(i, { stdin: e.target.value })}
                  rows={3}
                  className="font-mono text-xs"
                  placeholder="2 3"
                />
              </div>
              <div className="flex flex-col gap-1">
                <Label className="text-xs text-[var(--color-subtext0)]">期待出力 (expected)</Label>
                <Textarea
                  value={tc.expected_stdout}
                  onChange={(e) => updateTestCase(i, { expected_stdout: e.target.value })}
                  rows={3}
                  className="font-mono text-xs"
                  placeholder="5"
                />
              </div>
            </div>

            <div className="flex items-center gap-6">
              <label className="flex items-center gap-2 cursor-pointer text-xs text-[var(--color-text)]">
                <Checkbox
                  checked={tc.is_sample}
                  onCheckedChange={(v) => updateTestCase(i, { is_sample: v === true })}
                />
                サンプル表示
              </label>
              <div className="flex items-center gap-2">
                <span className="text-xs text-[var(--color-subtext0)]">スコア重み</span>
                <Input
                  type="number"
                  min={1}
                  value={tc.score_weight}
                  onChange={(e) =>
                    updateTestCase(i, { score_weight: Math.max(1, Number(e.target.value)) })
                  }
                  className="w-16 text-xs h-7"
                />
              </div>
            </div>
          </div>
        ))}
      </div>

      {error && (
        <div className="text-xs text-[var(--color-red)] px-3 py-2 rounded border border-[var(--color-red)] bg-[color-mix(in_srgb,var(--color-red)_10%,transparent)]">
          {error}
        </div>
      )}

      <Button type="submit" disabled={loading} className="self-start">
        {loading ? "投稿中..." : "問題を投稿する"}
      </Button>
    </form>
  );
}
