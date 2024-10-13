CREATE TABLE surveys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    created_at TEXT,
    parse_status TEXT,
    delivery_status TEXT,
    error_log TEXT,
    name TEXT NOT NULL UNIQUE,
    url_slug TEXT NOT NULL UNIQUE,
    config TEXT
);

CREATE TABLE surveys_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    created_at TEXT,
    completed_at TEXT,
    status TEXT,
    survey_id INTEGER NOT NULL,
    ip_addr TEXT,
    FOREIGN KEY (survey_id) REFERENCES surveys (id) ON DELETE CASCADE
);

CREATE TABLE surveys_questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    survey_id INTEGER NOT NULL,
    question_id TEXT NOT NULL,
    FOREIGN KEY (survey_id) REFERENCES surveys (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX surveys_questions_id ON surveys_questions (survey_id, question_id);

CREATE TABLE surveys_answers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    created_at TEXT,
    session_id INTEGER NOT NULL,
    question_id INTEGER NOT NULL,
    answer TEXT,
    FOREIGN KEY (session_id) REFERENCES surveys_sessions (id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES surveys_questions (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX surveys_answers_unique ON surveys_answers (session_id, question_id);
