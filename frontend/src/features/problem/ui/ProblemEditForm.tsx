"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { updateProblem } from "../api/updateProblem";
import type { Problem } from "../model/problem";
import { isEffectivelyActive } from "../model/problem";

// 公開設定の3状態
type PublishMode = "dated" | "open" | "hidden";

function toInitialMode(problem: Problem): PublishMode {
  if (problem.publishedAt) return "dated";
  if (problem.isActive) return "open";
  return "hidden";
}

type Props = {
  problem: Problem;
};

export default function ProblemEditForm({ problem }: Props) {
  const router = useRouter();

  const [title, setTitle] = useState(problem.title);
  const [description, setDescription] = useState(problem.description);
  const [publishMode, setPublishMode] = useState<PublishMode>(toInitialMode(problem));
  const [publishedAt, setPublishedAt] = useState(problem.publishedAt ?? "");

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit: React.EventHandler<React.SyntheticEvent<HTMLFormElement>> = async (e) => {
    e.preventDefault();
    if (publishMode === "dated" && !publishedAt) {
      setError("公開日を入力してください");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      await updateProblem(problem.id, {
        title,
        description,
        is_active: publishMode === "open",
        published_at: publishMode === "dated" && publishedAt ? publishedAt : null,
      });
      router.push("/admin/problems");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "不明なエラー");
    } finally {
      setLoading(false);
    }
  };

  const MODES: { value: PublishMode; label: string; description: string }[] = [
    {
      value: "dated",
      label: "公開日時指定",
      description: "指定した日付に今日の問題として表示されます",
    },
    {
      value: "open",
      label: "常時公開",
      description: "日付に関係なく問題一覧に常時表示されます",
    },
    {
      value: "hidden",
      label: "非公開",
      description: "公開一覧には表示されません",
    },
  ];

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6 max-w-2xl w-full">

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="title">タイトル</Label>
        <Input
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
        />
      </div>

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="description">問題文（Markdown）</Label>
        <Textarea
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={12}
          className="font-mono text-xs"
        />
      </div>

      {/* 公開設定 */}
      <fieldset className="flex flex-col gap-3 p-4 rounded-lg border border-[var(--color-surface0)] bg-[var(--color-mantle)]">
        <legend className="text-sm font-semibold text-[var(--color-text)] px-1">公開設定</legend>

        <div className="flex flex-col gap-2">
          {MODES.map((mode) => (
            <label
              key={mode.value}
              className={[
                "flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors",
                publishMode === mode.value
                  ? "border-[var(--color-mauve)] bg-[color-mix(in_srgb,var(--color-mauve)_8%,transparent)]"
                  : "border-[var(--color-surface1)] hover:border-[var(--color-surface2)]",
              ].join(" ")}
            >
              <input
                type="radio"
                name="publishMode"
                value={mode.value}
                checked={publishMode === mode.value}
                onChange={() => setPublishMode(mode.value)}
                className="mt-0.5 accent-[var(--color-mauve)]"
              />
              <div className="flex flex-col gap-0.5">
                <span className="text-sm font-medium text-[var(--color-text)]">{mode.label}</span>
                <span className="text-xs text-[var(--color-subtext0)]">{mode.description}</span>
              </div>
            </label>
          ))}
        </div>

        {publishMode === "dated" && (
          <div className="flex flex-col gap-1.5 pl-1">
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
      </fieldset>

      {/* 現在の状態プレビュー */}
      <p className="text-xs text-[var(--color-subtext0)]">
        保存後の状態:{" "}
        <span className="text-[var(--color-text)] font-medium">
          {publishMode === "dated" && publishedAt
            ? `${publishedAt} に自動公開`
            : publishMode === "dated"
            ? "日付を選択してください"
            : publishMode === "open"
            ? "常時公開中"
            : "非公開"}
        </span>
      </p>

      {error && (
        <div className="text-xs text-[var(--color-red)] px-3 py-2 rounded border border-[var(--color-red)] bg-[color-mix(in_srgb,var(--color-red)_10%,transparent)]">
          {error}
        </div>
      )}

      <div className="flex gap-3">
        <Button type="submit" disabled={loading}>
          {loading ? "保存中..." : "保存する"}
        </Button>
        <Button
          type="button"
          variant="secondary"
          onClick={() => router.push("/admin/problems")}
        >
          キャンセル
        </Button>
      </div>
    </form>
  );
}
