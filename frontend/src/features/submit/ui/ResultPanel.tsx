import type { JudgeResponse } from "../api/submitCode";
import { Badge } from "@/components/ui/badge";

type Props = {
  response: JudgeResponse;
};

export default function ResultPanel({ response }: Props) {
  return (
    <div className="flex items-start gap-4 flex-wrap">
      <Badge
        variant={response.overall === "Pass" ? "success" : "destructive"}
        className="text-sm px-3 py-1"
      >
        {response.overall === "Pass" ? "✓ 正解" : "✗ 不正解"}
      </Badge>

      <div className="flex gap-2 flex-wrap">
        {response.results.map((r) =>
          r.status === "Pass" ? (
            <div
              key={r.case_index}
              className="flex items-center gap-2 px-2.5 py-1 rounded text-xs bg-[color-mix(in_srgb,var(--color-green)_12%,transparent)] border border-[var(--color-green)] text-[var(--color-green)]"
            >
              <span className="font-semibold">Case {r.case_index}: Pass</span>
              <span className="text-[var(--color-overlay0)]">{r.time_ms.toFixed(0)}ms</span>
            </div>
          ) : (
            <div
              key={r.case_index}
              className="flex flex-col gap-1 px-2.5 py-1.5 rounded text-xs bg-[color-mix(in_srgb,var(--color-red)_12%,transparent)] border border-[var(--color-red)] text-[var(--color-red)]"
            >
              <div className="flex items-center gap-2">
                <span className="font-semibold">Case {r.case_index}: Fail</span>
                <span className="text-[var(--color-overlay0)]">{r.time_ms.toFixed(0)}ms</span>
              </div>
              <div className="font-mono text-[var(--color-subtext0)] space-y-0.5">
                {r.stderr ? (
                  <span>stderr: {r.stderr}</span>
                ) : (
                  <>
                    <div>出力: {JSON.stringify(r.stdout.trim())}</div>
                    <div>期待: {JSON.stringify(r.expected.trim())}</div>
                  </>
                )}
              </div>
            </div>
          )
        )}
      </div>
    </div>
  );
}
