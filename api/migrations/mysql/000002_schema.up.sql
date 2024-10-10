CREATE TABLE
    `surveys_sessions` (
        `id` INTEGER PRIMARY KEY AUTO_INCREMENT,
        `uuid` VARCHAR(36) NOT NULL UNIQUE DEFAULT (UUID()),
        `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
        `completed_at` DATETIME(3),
        `status` VARCHAR(255),
        `survey_id` INTEGER NOT NULL,
        `ip_addr` TEXT,
        FOREIGN KEY (`survey_id`) REFERENCES `surveys` (`id`) ON DELETE CASCADE
    );
