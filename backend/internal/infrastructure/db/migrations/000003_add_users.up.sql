-- ユーザー本体（プロバイダー非依存）
CREATE TABLE users (
    id         uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    email      varchar(255) NOT NULL UNIQUE,
    name       varchar(100) NOT NULL,
    role       varchar(20)  NOT NULL DEFAULT 'student',
    created_at timestamptz  NOT NULL DEFAULT NOW(),
    updated_at timestamptz  NOT NULL DEFAULT NOW()
);

-- OAuth プロバイダーとのひも付け（1ユーザーが複数プロバイダーを持てる）
CREATE TABLE user_identities (
    id           uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider     varchar(30)  NOT NULL,   -- 'google' | 'keycloak'
    provider_sub varchar(255) NOT NULL,   -- sub claim (プロバイダー固有 ID)
    created_at   timestamptz  NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_sub)
);

-- problems.created_by に FK 追加（NULL 許容のまま・既存データ保護）
ALTER TABLE problems
    ADD CONSTRAINT fk_problems_created_by
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

-- submissions.user_id に FK 追加（NULL 許容のまま）
ALTER TABLE submissions
    ADD CONSTRAINT fk_submissions_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
