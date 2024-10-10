CREATE TABLE
    `surveys` (
        `id` INTEGER PRIMARY KEY AUTO_INCREMENT,
        `uuid` VARCHAR(36) NOT NULL UNIQUE DEFAULT (UUID()),
        `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
        `parse_status` VARCHAR(255),
        `delivery_status` TEXT,
        `error_log` TEXT,
        `name` VARCHAR(255) NOT NULL UNIQUE,
        `url_slug` VARCHAR(255) NOT NULL UNIQUE,
        `config` TEXT
    );
