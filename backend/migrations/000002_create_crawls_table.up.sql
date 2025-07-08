CREATE TABLE crawls (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    url_id BIGINT UNSIGNED NOT NULL,
    status ENUM('queued', 'running', 'completed', 'error') DEFAULT 'queued',
    started_at TIMESTAMP NULL DEFAULT NULL,
    completed_at TIMESTAMP NULL DEFAULT NULL,
    error_message TEXT DEFAULT '',
    internal_links INT UNSIGNED DEFAULT 0,
    external_links INT UNSIGNED DEFAULT 0,
    broken_links INT UNSIGNED DEFAULT 0,
    heading_counts JSON DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE,
    INDEX idx_crawls_url_id (url_id),
    INDEX idx_crawls_status (status),
    INDEX idx_crawls_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci; 