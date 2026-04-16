ALTER TABLE submissions DROP CONSTRAINT IF EXISTS fk_submissions_user;
ALTER TABLE problems    DROP CONSTRAINT IF EXISTS fk_problems_created_by;
DROP TABLE IF EXISTS user_identities;
DROP TABLE IF EXISTS users;
