# フロントエンド画面設計

## 決定事項（Q1〜Q5）

| # | 質問 | 決定 |
|---|------|------|
| Q1 | 今日の問題がない日の `/` の挙動 | メッセージを表示し `/problems`（一覧）へのリンクを出す。自動リダイレクトはしない |
| Q2 | 提出コードの公開範囲 | Phase2（ユーザー実装）まで提出詳細ページは作らない。スコアと Pass/Fail のみ画面内表示 |
| Q3 | ナビゲーション構造 | 共通 `Header` に `問題一覧` を常時表示。`admin` ロール保持時のみ `問題管理` / `問題作成` を追加表示 |
| Q4 | 管理画面の認証 | Phase2 まで認証なし（`/admin/*` は無保護） |
| Q5 | 非サンプルテストケースの stdin 表示 | **非表示**。`is_sample = false` のテストケースの入出力は提出詳細でも見せない |

---

## 公開モード設計

問題には2種類の公開モードがある。どちらも `is_active = true` のとき一般ユーザーが解答できる。

| モード | `published_at` | 説明 |
|--------|---------------|------|
| 日付指定 | `YYYY-MM-DD` | その日の「今日の問題」として `/` に表示される |
| 常時公開 | `null` | 日付によらず `/problems` 一覧に常時掲載される |

---

## 画面一覧・URI 設計（確定）

### 公開画面

| # | URI | 画面名 | 概要 |
|---|-----|--------|------|
| 1 | `/` | 今日の問題 | JST 当日の `published_at` を持つ問題を表示。なければ「今日の問題はまだ公開されていません」＋一覧リンク |
| 2 | `/problems` | 問題アーカイブ | `is_active = true` の全問題を `published_at` 降順で一覧表示 |
| 3 | `/problems/[id]` | 問題詳細・解答 | 任意の問題を解く画面。`/` と同レイアウト |
| 4 | `/problems/[id]/submissions` | 提出履歴 | **Phase2 以降実装**。スコア・判定の一覧（コードは非公開） |
| 5 | `/problems/[id]/submissions/[subId]` | 提出詳細 | **Phase2 以降実装**。非サンプルの入出力は非表示 |

### 管理画面

| # | URI | 画面名 | 概要 |
|---|-----|--------|------|
| 6 | `/admin/problems` | 問題管理一覧 | 全問題（非公開含む）を一覧。公開モード（日付指定 / 常時公開）をバッジで表示 |
| 7 | `/admin/problems/new` | 問題作成 | タイトル・問題文・公開モード（日付指定 or 常時公開）・テストケースを入力して投稿 |
| 8 | `/admin/problems/[id]/edit` | 問題編集 | **Phase2 以降実装** |

---

## バックエンド API 一覧

### 実装済み

| メソッド | パス | 用途 |
|---------|------|------|
| `GET` | `/api/problems` | 公開中問題の一覧（`is_active = true`） |
| `GET` | `/api/problems/today` | JST 当日の問題を1件返す。なければ 404 |
| `GET` | `/api/problems/:id` | 問題1件取得 |
| `POST` | `/api/problems` | 問題作成（テストケース同時登録） |
| `POST` | `/api/submit` | コード提出・採点 |

### Phase2 以降で追加予定

| メソッド | パス | 用途 |
|---------|------|------|
| `PUT` | `/api/problems/:id` | 問題更新 |
| `GET` | `/api/problems/:id/submissions` | 問題への提出一覧 |
| `GET` | `/api/submissions/:id` | 提出詳細（非サンプルの stdin は返さない） |
| `GET` | `/api/admin/problems` | 管理用：非公開含む全問題一覧 |

---

## モックユーザー（Phase2 まで）

`src/shared/model/user.ts` に `MOCK_USER` を定義。`UserProvider` で React Context を通じてアプリ全体に供給する。

```ts
// role を "student" に変えると管理者リンクが非表示になる
export const MOCK_USER: User = {
  id: "00000000-0000-0000-0000-000000000001",
  name: "テスト管理者",
  email: "admin@linuxclub.example",
  role: "admin",  // "student" | "admin"
};
```

Phase2 では `UserProvider` 内で Keycloak トークンを検証して差し替える。
