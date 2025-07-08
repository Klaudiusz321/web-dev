CREATE TABLE links (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    url_id BIGINT UNSIGNED NOT NULL,
    crawl_id BIGINT UNSIGNED NOT NULL,
    link_url VARCHAR(2048) NOT NULL,
    link_text VARCHAR(512) DEFAULT '',
    link_type ENUM('internal', 'external') NOT NULL,
    status_code INT UNSIGNED DEFAULT 0,
    is_accessible BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE,
    FOREIGN KEY (crawl_id) REFERENCES crawls(id) ON DELETE CASCADE,
    INDEX idx_links_url_id (url_id),
    INDEX idx_links_crawl_id (crawl_id),
    INDEX idx_links_type (link_type),
    INDEX idx_links_accessible (is_accessible),
    INDEX idx_links_status_code (status_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci; 