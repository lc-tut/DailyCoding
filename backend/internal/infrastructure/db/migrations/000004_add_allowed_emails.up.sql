-- ドメイン制限に加えて個別に許可するメールアドレスを管理するテーブル
CREATE TABLE allowed_emails (
    email      varchar(255) PRIMARY KEY,
    note       text,
    created_at timestamptz NOT NULL DEFAULT NOW()
);
