-- submissions.user_id を NOT NULL に変更（認証必須化）
-- ON DELETE SET NULL を ON DELETE CASCADE に変更し FK を張り直す
ALTER TABLE submissions
    DROP CONSTRAINT fk_submissions_user;

ALTER TABLE submissions
    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE submissions
    ADD CONSTRAINT fk_submissions_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
