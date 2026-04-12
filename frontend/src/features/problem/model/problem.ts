export type Problem = {
  id: string;
  title: string;
  description: string; // Markdown
};

// MVP: 問題はハードコード。DB追加後はAPIから取得する想定
export const currentProblem: Problem = {
  id: "sum-two-numbers",
  title: "二数の和",
  description: `## 問題：二数の和

標準入力から空白区切りで 2 つの整数を受け取り、その和を出力してください。

### 入力形式

\`\`\`
A B
\`\`\`

### 出力形式

\`\`\`
A + B の値
\`\`\`

### サンプル

| 入力 | 出力 |
|------|------|
| \`2 3\` | \`5\` |
| \`10 20\` | \`30\` |
`,
};

export const initialCode = `a, b = map(int, input().split())
print(a + b)
`;
