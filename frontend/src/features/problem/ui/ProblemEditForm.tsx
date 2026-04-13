"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { updateProblem } from "../api/updateProblem";
import type { Problem } from "../model/problem";

type PublishMode = "dated" | "open";

type Props = {
  problem: Problem;
};

export default function ProblemEditForm({ problem }: Props) {
  const router = useRouter();

  const [title, setTitle] = useState(problem.title);
  const [description, setDescription] = useState(problem.description);
  const [isActive, setIsActive] = useState(problem.isActive);
  const [publishMode, setPublishMode] = useState<PublishMode>(
    problem.publishedAt ? "dated" : "open"
  );
  const [publishedAt, setPublishedAt] = useState(problem.publishedAt ?? "");

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit: React.EventHandler<React.SyntheticEvent<HTMLFormElement>> = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await updateProblem(problem.id, {
        title,
        description,
        is_active: isActive,
        published_at: publishMode === "dated" && publishedAt ? publishedAt : null,
      });
      router.push("/admin/problems");
      router.refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "不明なエラー");
    } finally {
      setLoading(false);
    }
  };

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
            公開する（is_active = true）
          </Label>
        </div>
      </fieldset>

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
