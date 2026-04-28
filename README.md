# DaileyCoding

LinuxClub の毎日 Python コーディング練習サービス。

## Stack

| レイヤー | 技術 |
|---|---|
| Frontend | Next.js 16 (App Router) / TypeScript / Tailwind CSS v4 |
| Backend | Go / Echo v5 |
| DB | PostgreSQL 17 |
| コード実行 | Docker (`python:3.14-slim`) / snekbox (nsjail) |
| 認証 | OAuth2 Authorization Code Flow + HttpOnly Cookie (JWT) |
| バージョン管理 | mise |

---

## 前提条件

- [mise](https://mise.jdx.dev/) がインストールされていること
- Docker / Docker Compose が動作していること

---

## セットアップ

### 1. ツールのインストール

```bash
mise install
```

Go・Node.js・pnpm が `mise.toml` に記載のバージョンで揃います。

### 2. フロントエンド依存関係のインストール

```bash
cd frontend && pnpm install && cd ..
```

### 3. 環境変数の設定

バックエンドは起動時に環境変数を読みます。ローカル開発は `.env` ファイルか shell で設定してください。

#### コード実行エンジンの選択

| 環境変数 | 値 | 説明 |
|---|---|---|
| `EXECUTOR_TYPE` | `docker`（デフォルト） | Docker コンテナで Python を実行 |
| `EXECUTOR_TYPE` | `snekbox` | snekbox (nsjail) 経由で Python を実行 |
| `SNEKBOX_URL` | `http://snekbox:8060`（デフォルト） | snekbox サービスの URL |

snekbox を使う場合は `docker-compose.snekbox.yml` も参照してください（後述）。

#### 最小構成（認証バイパスあり）

```bash
cp backend/.env.example backend/.env
```

> **`AUTH_BYPASS` について**
> `AUTH_BYPASS=true` にすると JWT 検証を完全スキップし、固定のテスト管理者として動作します。
> **デフォルトは `false`** です。開発時も OIDC プロバイダーを使ってください。

#### Google OAuth2 を使う場合

[Google Cloud Console](https://console.cloud.google.com/) でプロジェクトを作成し、OAuth 2.0 クライアント ID を発行します。

- 承認済みリダイレクト URI: `http://localhost:8080/api/auth/google/callback`

```bash
export GOOGLE_CLIENT_ID="your-client-id.apps.googleusercontent.com"
export GOOGLE_CLIENT_SECRET="your-client-secret"
export GOOGLE_CALLBACK_URL="http://localhost:8080/api/auth/google/callback"
export JWT_SECRET="任意のランダム文字列（本番は強いシークレットを使うこと）"
export AUTH_BYPASS=false
```

#### Keycloak を使う場合

```bash
export KEYCLOAK_CLIENT_ID="daileycoding"
export KEYCLOAK_CLIENT_SECRET="your-secret"
export KEYCLOAK_ISSUER="http://keycloak:8080/realms/daileycoding"
export KEYCLOAK_CALLBACK_URL="http://localhost:8080/api/auth/keycloak/callback"
export JWT_SECRET="任意のランダム文字列"
export AUTH_BYPASS=false
```

### 4. DB の起動とマイグレーション

```bash
mise run db          # PostgreSQL コンテナを起動
mise run db:migrate  # マイグレーションを実行（初回・変更時）
```

マイグレーションは `backend/internal/infrastructure/db/migrations/` 以下の SQL ファイルが順番に適用されます。

### 5. 管理者アカウントの登録

ログインすると `role=student` のユーザーが自動作成されます。
管理者への昇格には **2 つの方法** があります。

#### 方法 A: DB を直接更新（手動・初回セットアップ向け）

まず対象ユーザーが一度ログインしてアカウントを作成してから実行します。

```bash
docker compose exec postgres psql -U daileycoding -d daileycoding
```

```sql
-- ログイン済みユーザーを管理者に昇格
UPDATE users SET role = 'admin' WHERE email = 'yourname@example.com';

-- 昇格後は再ログインすると admin ロールが反映される
```

#### 方法 B: authentik グループで自動同期（推奨・本番環境向け）

authentik を OIDC プロバイダーとして使う場合、グループ単位で管理者を自動付与できます。

**authentik 側の設定:**

1. authentik 管理画面でグループ（例: `dailycoding-admins`）を作成し、管理者にしたいユーザーを追加する
2. Provider の **Advanced Protocol Settings → Scopes** に `groups` Property Mapping を追加する
   （これにより ID Token に `groups` クレームが含まれるようになる）

**バックエンド側の設定:**

```bash
# backend/.env.prod に追記
ADMIN_GROUPS=dailycoding-admins   # カンマ区切りで複数指定可
```

この設定を有効にすると、ログイン時にグループが照合されロールが自動で同期されます。

| 状態 | 結果 |
|---|---|
| グループに所属している | `role = admin` |
| グループから除外された | 次回ログイン時に `role = student` に戻る |
| `ADMIN_GROUPS` 未設定 | ロールを変更しない（方法 A の手動管理） |

管理者になると `/admin/problems` の問題管理ページにアクセスできます。

### 6. 外部メールアドレスの個別許可

`ALLOWED_DOMAINS` に含まれないドメインのユーザー（外部協力者など）を個別に許可できます。

```bash
docker compose exec postgres psql -U daileycoding -d daileycoding
```

```sql
-- 外部メールアドレスを許可リストに追加
INSERT INTO allowed_emails (email, note)
VALUES ('guest@gmail.com', '外部協力者 山田太郎');

-- 一覧確認
SELECT * FROM allowed_emails;

-- 削除（アクセス停止）
DELETE FROM allowed_emails WHERE email = 'guest@gmail.com';
```

> `ALLOWED_DOMAINS` が未設定の場合はこのテーブルに関係なく全員許可になります。
> 本番環境では必ず `ALLOWED_DOMAINS` を設定してください。

---

## 起動

ターミナルを 3 つ用意して、それぞれ実行します。

```bash
# 1. DB（起動済みであればスキップ）
mise run db

# 2. バックエンド（port 8080）
mise run backend

# 3. フロントエンド（port 3000）
mise run frontend
```

ブラウザで `http://localhost:3000` を開きます。

### その他の mise タスク

| コマンド | 内容 |
|---|---|
| `mise run db` | PostgreSQL コンテナを起動 |
| `mise run db:migrate` | 未適用マイグレーションを実行 |
| `mise run db:reset` | データ全削除 → DB 再起動 → マイグレーション |
| `mise run backend` | Go API サーバー起動 (port 8080) |
| `mise run frontend` | Next.js 開発サーバー起動 (port 3000) |
| `mise run snekbox:dev` | snekbox をローカル開発用に起動（ポート 8060 をホストに公開） |
| `mise run snekbox:dev:down` | snekbox 開発コンテナを停止 |
| `mise run snekbox:up` | snekbox executor で本番スタックを起動 |
| `mise run snekbox:down` | snekbox を含む全サービスを停止 |
| `mise run snekbox:logs` | snekbox のログをリアルタイム表示 |

### snekbox 実行エンジン（オプション）

[snekbox](https://github.com/python-discord/snekbox) は nsjail によるシステムコールレベルのサンドボックスを提供します。
Docker executor より強固なコンテナエスケープ対策が必要な場合に使用します。

| 項目 | Docker executor | snekbox |
|---|---|---|
| 分離方式 | Docker コンテナ | nsjail (seccomp-bpf + namespaces) |
| メモリ制限 | 64MB | snekbox 設定で制御 |
| 要 privileged | 不要 | snekbox コンテナに必要（nsjail の要件） |

#### ローカル開発環境で snekbox を使う

開発時は Go バックエンドがホストで動くため、snekbox のポートをホストに公開します。

```bash
# 1. snekbox コンテナを起動（ポート 8060 をホストに公開）
mise run snekbox:dev

# 2. backend/.env に以下を追加
EXECUTOR_TYPE=snekbox
SNEKBOX_URL=http://localhost:8060

# 3. バックエンドを通常通り起動
mise run backend
```

停止するときは:

```bash
mise run snekbox:dev:down
```

#### 本番環境で snekbox を使う

```bash
mise run snekbox:up    # snekbox + バックエンドをオーバーライドして起動
mise run snekbox:down  # snekbox を含む全サービスを停止
```

---

## 認証フロー概要

```
ブラウザ
  → GET /api/auth/google          # state Cookie を発行し Google 認可画面へリダイレクト
  → Google 同意画面
  → GET /api/auth/google/callback  # コード交換 → ユーザー情報取得 → JIT プロビジョニング
  → HttpOnly Cookie (dc_session) に JWT を発行
  → フロントエンドへリダイレクト
```

- Cookie 名: `dc_session`（HttpOnly / SameSite=Lax / 有効期限 24 時間）
- JWT ペイロード: `uid` / `email` / `name` / `role`
- 複数プロバイダー対応: 同じメールアドレスで Google と Keycloak の両方からログインすると同一ユーザーに紐づく

---

## ディレクトリ構成

```
.
├── mise.toml              # ツールバージョン・タスク定義
├── docker-compose.yml          # PostgreSQL
├── docker-compose.snekbox.yml      # snekbox 本番用オーバーライド
├── docker-compose.snekbox.dev.yml  # snekbox 開発用オーバーライド（ポート公開）
├── backend/
│   ├── cmd/server/        # エントリポイント
│   ├── internal/
│   │   ├── domain/        # エンティティ・型定義
│   │   ├── handler/       # HTTP ハンドラー
│   │   ├── middleware/    # JWT 認証ミドルウェア
│   │   ├── usecase/       # ビジネスロジック
│   │   └── infrastructure/
│   │       ├── db/        # PostgreSQL（migrations・repository）
│   │       ├── docker/    # Docker コード実行エンジン
│   │       └── snekbox/   # snekbox (nsjail) コード実行エンジン
└── frontend/
    └── src/
        ├── app/           # Next.js App Router ページ
        ├── features/      # 機能ごとのコンポーネント・API
        ├── shared/        # 共通 Provider・型・設定
        └── components/    # 汎用 UI コンポーネント
```
