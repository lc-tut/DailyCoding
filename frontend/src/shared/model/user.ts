export type UserRole = "student" | "admin";

export type User = {
  id: string;
  name: string;
  email: string;
  role: UserRole;
};

/**
 * デバッグ用モックユーザー。
 * Phase2 で Keycloak 認証に置き換える。
 * ロールを切り替えたい場合は role を "student" に変更する。
 */
export const MOCK_USER: User = {
  id: "00000000-0000-0000-0000-000000000001",
  name: "テスト管理者",
  email: "admin@linuxclub.example",
  role: "admin",
};
