import type { JudgeResponse } from "../api/submitCode";
import styles from "./ResultPanel.module.css";

type Props = {
  response: JudgeResponse;
};

// ResultPanel は "use client" 境界内（SubmissionArea）から呼ばれるため
// それ自体に "use client" は不要
export default function ResultPanel({ response }: Props) {
  return (
    <div className={styles.wrapper}>
      <span
        className={
          response.overall === "Pass" ? styles.overallPass : styles.overallFail
        }
      >
        {response.overall === "Pass" ? "✓ 正解" : "✗ 不正解"}
      </span>

      <div className={styles.cases}>
        {response.results.map((r) => (
          <div
            key={r.case_index}
            className={r.status === "Pass" ? styles.casePass : styles.caseFail}
          >
            <span className={styles.caseLabel}>
              Case {r.case_index}: {r.status}
            </span>
            <span className={styles.caseTime}>{r.time_ms.toFixed(0)}ms</span>

            {r.status === "Fail" && (
              <div className={styles.diff}>
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
  );
}
