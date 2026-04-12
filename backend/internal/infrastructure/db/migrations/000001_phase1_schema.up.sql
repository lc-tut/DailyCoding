CREATE TABLE problems (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title        varchar(200) NOT NULL,
    description  text         NOT NULL,
    created_by   uuid         NULL,     -- Phase2 で users FK を追加予定
    is_active    boolean      NOT NULL DEFAULT false,
    published_at date         NULL,
    created_at   timestamptz  NOT NULL DEFAULT NOW(),
    updated_at   timestamptz  NOT NULL DEFAULT NOW()
);

CREATE TABLE test_cases (
    id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id       uuid        NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    stdin            text        NOT NULL,
    expected_stdout  text        NOT NULL,
    order_index      smallint    NOT NULL,
    is_sample        boolean     NOT NULL DEFAULT false,
    hint             text        NULL,
    score_weight     smallint    NOT NULL DEFAULT 1,
    created_at       timestamptz NOT NULL DEFAULT NOW(),
    updated_at       timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (problem_id, order_index)
);

CREATE TABLE submissions (
    id             uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        uuid        NULL,     -- Phase2 で NOT NULL + FK に変更予定
    problem_id     uuid        NOT NULL REFERENCES problems(id),
    code           text        NOT NULL,
    language       varchar(30) NOT NULL DEFAULT 'python',
    overall_status varchar(10) NOT NULL,
    score          smallint    NOT NULL DEFAULT 0,
    ai_review      text        NULL,
    created_at     timestamptz NOT NULL DEFAULT NOW(),
    updated_at     timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_submissions_ranking
    ON submissions (problem_id, score DESC, created_at ASC);

CREATE TABLE submission_results (
    id                uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id     uuid    NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    test_case_id      uuid    NOT NULL REFERENCES test_cases(id),
    status            varchar(10) NOT NULL,
    stdout            text    NOT NULL,
    stderr            text    NOT NULL,
    execution_time_ms real    NOT NULL,
    created_at        timestamptz NOT NULL DEFAULT NOW(),
    updated_at        timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (submission_id, test_case_id)
);
