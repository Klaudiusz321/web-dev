-- Create the webcrawler database if it doesn't exist
CREATE DATABASE IF NOT EXISTS webcrawler CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Use the webcrawler database
USE webcrawler;

-- Create the urls table
CREATE TABLE IF NOT EXISTS urls (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    url VARCHAR(2048) NOT NULL UNIQUE,
    title VARCHAR(512) DEFAULT '',
    html_version VARCHAR(50) DEFAULT '',
    status ENUM('pending', 'running', 'completed', 'error') DEFAULT 'pending',
    has_login_form BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    INDEX idx_urls_status (status),
    INDEX idx_urls_created_at (created_at),
    INDEX idx_urls_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create the crawls table
CREATE TABLE IF NOT EXISTS crawls (
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

-- Create the links table
CREATE TABLE IF NOT EXISTS links (
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