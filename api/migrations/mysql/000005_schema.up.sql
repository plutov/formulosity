CREATE TABLE
    `surveys_answers` (
        `id` INTEGER PRIMARY KEY AUTO_INCREMENT,
        `uuid` VARCHAR(36) NOT NULL UNIQUE DEFAULT (UUID()),
        `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
        `session_id` INTEGER NOT NULL,
        `question_id` INTEGER NOT NULL,
        `answer` TEXT,
        FOREIGN KEY (`session_id`) REFERENCES surveys_sessions (`id`) ON DELETE CASCADE,
        FOREIGN KEY (`question_id`) REFERENCES surveys_questions (`id`) ON DELETE CASCADE
    );
