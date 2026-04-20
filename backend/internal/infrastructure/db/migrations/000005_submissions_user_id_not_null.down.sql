-- submissions.user_id を NULL 許容に戻す
ALTER TABLE submissions
    DROP CONSTRAINT fk_submissions_user;

ALTER TABLE submissions
    ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE submissions
    ADD CONSTRAINT fk_submissions_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
