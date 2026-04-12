import ProblemPanel from "@/features/problem/ui/ProblemPanel";
import SubmissionArea from "@/features/submit/ui/SubmissionArea";
import { currentProblem, initialCode } from "@/features/problem/model/problem";
import styles from "./page.module.css";

// サーバーコンポーネント: 静的な問題データを子へ渡す
export default function Home() {
  return (
    <div className={styles.layout}>
      <header className={styles.header}>
        <span className={styles.logo}>DaileyCoding</span>
        <span className={styles.subtitle}>LinuxClub Python Challenge</span>
      </header>

      <main className={styles.main}>
        <ProblemPanel problem={currentProblem} />
        <SubmissionArea initialCode={initialCode} />
      </main>
    </div>
  );
}
