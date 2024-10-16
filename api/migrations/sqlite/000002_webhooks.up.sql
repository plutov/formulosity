CREATE TABLE surveys_webhook_responses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TEXT,
    session_id INTEGER NOT NULL,
    response_status INTEGER NOT NULL,
    response TEXT,
    FOREIGN KEY (session_id) REFERENCES surveys_sessions (id) ON DELETE CASCADE
);