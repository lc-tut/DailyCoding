/**
 * 認証が必要なパスかどうかを判定する。
 * ログアウト後にこれらのページにいる場合はホームへリダイレクトする。
 */
const AUTH_REQUIRED_PATTERNS: Array<(path: string) => boolean> = [
  (path) => path.startsWith("/admin"),
  (path) => path === "/problems/new",
];

export function requiresAuth(pathname: string): boolean {
  return AUTH_REQUIRED_PATTERNS.some((match) => match(pathname));
}
