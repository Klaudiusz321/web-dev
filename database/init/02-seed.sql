-- Use the webcrawler database
USE webcrawler;

-- Insert sample URLs for testing
INSERT IGNORE INTO urls (url, title, html_version, status, has_login_form, created_at) VALUES
('https://example.com', 'Example Domain', 'HTML5', 'completed', FALSE, NOW()),
('https://github.com', 'GitHub: Let\'s build from here', 'HTML5', 'completed', TRUE, NOW()),
('https://stackoverflow.com', 'Stack Overflow - Where Developers Learn, Share, & Build Careers', 'HTML5', 'pending', TRUE, NOW()),
('https://www.w3.org', 'W3C', 'HTML5', 'error', FALSE, NOW());

-- Insert sample crawl data
INSERT IGNORE INTO crawls (url_id, status, started_at, completed_at, internal_links, external_links, broken_links, heading_counts, created_at) VALUES
(1, 'completed', NOW() - INTERVAL 1 HOUR, NOW() - INTERVAL 55 MINUTE, 15, 8, 2, '{"h1":1,"h2":3,"h3":5,"h4":2,"h5":1,"h6":0}', NOW()),
(2, 'completed', NOW() - INTERVAL 2 HOUR, NOW() - INTERVAL 115 MINUTE, 42, 12, 0, '{"h1":2,"h2":8,"h3":12,"h4":5,"h5":3,"h6":1}', NOW()),
(4, 'error', NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 25 MINUTE, 0, 0, 0, NULL, NOW());

-- Insert sample links data
INSERT IGNORE INTO links (url_id, crawl_id, link_url, link_text, link_type, status_code, is_accessible, created_at) VALUES
-- Links for example.com
(1, 1, 'https://example.com/about', 'About Us', 'internal', 200, TRUE, NOW()),
(1, 1, 'https://example.com/contact', 'Contact', 'internal', 200, TRUE, NOW()),
(1, 1, 'https://example.com/broken-link', 'Broken Page', 'internal', 404, FALSE, NOW()),
(1, 1, 'https://www.iana.org/domains/example', 'More information...', 'external', 200, TRUE, NOW()),
(1, 1, 'https://broken-external.com/test', 'Broken External', 'external', 500, FALSE, NOW()),

-- Links for github.com
(2, 2, 'https://github.com/features', 'Features', 'internal', 200, TRUE, NOW()),
(2, 2, 'https://github.com/pricing', 'Pricing', 'internal', 200, TRUE, NOW()),
(2, 2, 'https://github.com/enterprise', 'Enterprise', 'internal', 200, TRUE, NOW()),
(2, 2, 'https://docs.github.com', 'Documentation', 'external', 200, TRUE, NOW()),
(2, 2, 'https://github.blog', 'Blog', 'external', 200, TRUE, NOW()),
(2, 2, 'https://education.github.com', 'Education', 'external', 200, TRUE, NOW()); 