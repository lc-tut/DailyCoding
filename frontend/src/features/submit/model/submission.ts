export type SubmissionSummary = {
  id: string;
  language: string;
  overallStatus: "Pass" | "Fail";
  score: number;
  createdAt: string; // ISO8601
};

export type TestCaseResult = {
  orderIndex: number;
  isSample: boolean;
  stdin: string; // "" if not sample
  expectedStdout: string; // "" if not sample
  status: "Pass" | "Fail";
  stdout: string;
  stderr: string;
  executionTimeMs: number;
};

export type SubmissionDetail = {
  id: string;
  problemId: string;
  code: string;
  language: string;
  overallStatus: "Pass" | "Fail";
  score: number;
  createdAt: string;
  results: TestCaseResult[];
};
