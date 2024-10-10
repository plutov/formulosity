CREATE TABLE
    `surveys_questions` (
        `id` INTEGER PRIMARY KEY AUTO_INCREMENT,
        `uuid` VARCHAR(36) NOT NULL UNIQUE DEFAULT (UUID()),
        `survey_id` INTEGER NOT NULL,
        `question_id` VARCHAR(255) NOT NULL,
        FOREIGN KEY (`survey_id`) REFERENCES `surveys` (`id`) ON DELETE CASCADE
    );
