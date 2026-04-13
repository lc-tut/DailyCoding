export type Problem = {
  id: string;
  title: string;
  description: string; // Markdown
  isActive: boolean;
  publishedAt: string | null;
};

// isEffectivelyActive は published_at が当日以前の場合も公開済みとみなす
export function isEffectivelyActive(p: Problem): boolean {
  if (p.isActive) return true;
  if (p.publishedAt) {
    const today = new Date().toISOString().split("T")[0];
    return p.publishedAt <= today;
  }
  return false;
}

export const initialCode = `a, b = map(int, input().split())
print(a + b)
`;
