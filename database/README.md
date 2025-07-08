# Database Schema and Migrations

This directory contains the MySQL database schema, migrations, and initialization scripts for the Web Crawler application.

## 📊 Database Schema Overview

### Tables Structure

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    URLs     │────▶│   Crawls    │────▶│   Links     │
│             │     │             │     │             │
│ id (PK)     │     │ id (PK)     │     │ id (PK)     │
│ url         │     │ url_id (FK) │     │ url_id (FK) │
│ title       │     │ status      │     │ crawl_id(FK)│
│ status      │     │ results...  │     │ link_url    │
│ ...         │     │ ...         │     │ ...         │
└─────────────┘     └─────────────┘     └─────────────┘
```

### 🗂 URLs Table
**Purpose**: Stores websites to be crawled

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGINT UNSIGNED PK | Auto-increment primary key |
| `url` | VARCHAR(2048) UNIQUE | The website URL to crawl |
| `title` | VARCHAR(512) | Page title extracted from `<title>` tag |
| `html_version` | VARCHAR(50) | HTML version (e.g., "HTML5") |
| `status` | ENUM | Current status: `pending`, `running`, `completed`, `error` |
| `has_login_form` | BOOLEAN | Whether a login form was detected |
| `created_at` | TIMESTAMP | When the URL was added |
| `updated_at` | TIMESTAMP | Last update time |
| `deleted_at` | TIMESTAMP | Soft delete timestamp (NULL if not deleted) |

**Indexes**:
- `idx_urls_status` - Query by status
- `idx_urls_created_at` - Sort by creation date
- `idx_urls_deleted_at` - Soft delete lookups

### 🔄 Crawls Table
**Purpose**: Stores individual crawling sessions

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGINT UNSIGNED PK | Auto-increment primary key |
| `url_id` | BIGINT UNSIGNED FK | Foreign key to `urls.id` |
| `status` | ENUM | Crawl status: `queued`, `running`, `completed`, `error` |
| `started_at` | TIMESTAMP | When crawling started |
| `completed_at` | TIMESTAMP | When crawling finished |
| `error_message` | TEXT | Error details if status is `error` |
| `internal_links` | INT UNSIGNED | Count of internal links found |
| `external_links` | INT UNSIGNED | Count of external links found |
| `broken_links` | INT UNSIGNED | Count of broken/inaccessible links |
| `heading_counts` | JSON | Count of heading tags `{"h1":1,"h2":3,...}` |
| `created_at` | TIMESTAMP | When the crawl record was created |
| `updated_at` | TIMESTAMP | Last update time |

**Indexes**:
- `idx_crawls_url_id` - Find crawls by URL
- `idx_crawls_status` - Filter by crawl status
- `idx_crawls_created_at` - Sort by creation date

### 🔗 Links Table
**Purpose**: Stores individual links found during crawling

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGINT UNSIGNED PK | Auto-increment primary key |
| `url_id` | BIGINT UNSIGNED FK | Foreign key to `urls.id` |
| `crawl_id` | BIGINT UNSIGNED FK | Foreign key to `crawls.id` |
| `link_url` | VARCHAR(2048) | The discovered link URL |
| `link_text` | VARCHAR(512) | Text content of the link |
| `link_type` | ENUM | Type: `internal` or `external` |
| `status_code` | INT UNSIGNED | HTTP status code when checking accessibility |
| `is_accessible` | BOOLEAN | Whether the link is accessible |
| `created_at` | TIMESTAMP | When the link was discovered |

**Indexes**:
- `idx_links_url_id` - Find links by URL
- `idx_links_crawl_id` - Find links by crawl session
- `idx_links_type` - Filter by link type
- `idx_links_accessible` - Filter by accessibility
- `idx_links_status_code` - Filter by HTTP status

## 🚀 Migration System

### Directory Structure
```
backend/
├── migrations/                 # golang-migrate files
│   ├── 000001_create_urls_table.up.sql
│   ├── 000001_create_urls_table.down.sql
│   ├── 000002_create_crawls_table.up.sql
│   ├── 000002_create_crawls_table.down.sql
│   ├── 000003_create_links_table.up.sql
│   └── 000003_create_links_table.down.sql
└── cmd/migrate/               # Migration CLI tool
    └── main.go
```

```
database/
└── init/                      # Docker initialization
    ├── 01-init.sql           # Schema creation
    └── 02-seed.sql           # Sample data
```

### Using Migrations

#### 1. CLI Migration Tool
```bash
cd backend

# Build the migration tool
make build-migrate

# Apply all pending migrations
make migrate-up

# Rollback one migration
make migrate-down

# Check current version
make migrate-version

# Reset all migrations (careful!)
make migrate-reset
```

#### 2. Manual Commands
```bash
# Apply migrations
./bin/migrate -action=up

# Rollback 2 steps
./bin/migrate -action=down -steps=2

# Check version
./bin/migrate -action=version
```

#### 3. Programmatic (Development)
The application automatically runs migrations on startup:
- **Production**: Uses file-based migrations (golang-migrate)
- **Development**: Falls back to GORM AutoMigrate if files fail

## 🐳 Docker Setup

### Automatic Initialization
When using Docker Compose, the database is automatically initialized with:

1. **Schema Creation** (`01-init.sql`): Creates all tables with proper indexes
2. **Sample Data** (`02-seed.sql`): Inserts test data for development

### Connection Details
```yaml
# docker-compose.yml
mysql:
  environment:
    MYSQL_ROOT_PASSWORD: rootpassword
    MYSQL_DATABASE: webcrawler
    MYSQL_USER: crawler
    MYSQL_PASSWORD: password
  ports:
    - "3306:3306"
```

### Database URL Format
```
root:password@tcp(localhost:3306)/webcrawler?charset=utf8mb4&parseTime=True&loc=Local
```

## 🛠 Development Commands

### Start MySQL with Docker
```bash
docker run --name mysql-webcrawler \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=webcrawler \
  -p 3306:3306 \
  -d mysql:8.0
```

### Connect to Database
```bash
# Using MySQL client
mysql -h localhost -u root -p webcrawler

# Using Docker exec
docker exec -it mysql-webcrawler mysql -u root -p webcrawler
```

### Backup and Restore
```bash
# Backup
mysqldump -h localhost -u root -p webcrawler > backup.sql

# Restore
mysql -h localhost -u root -p webcrawler < backup.sql
```

## 📈 Performance Considerations

### Indexes
- All foreign keys have indexes for fast JOINs
- Status columns are indexed for filtering
- Created_at columns are indexed for sorting
- Unique constraint on `urls.url` prevents duplicates

### Partitioning (Future)
For large datasets, consider partitioning by:
- `crawls` table by `created_at` (monthly partitions)
- `links` table by `url_id` or `created_at`

### Query Optimization
- Use `LIMIT` and `OFFSET` for pagination
- Filter by indexed columns when possible
- Use `EXISTS` instead of `IN` for large subqueries

## 🔒 Security Features

- **Soft Deletes**: URLs use `deleted_at` for recovery
- **Cascade Deletes**: Removing a URL removes all related crawls and links
- **UTF8MB4**: Full Unicode support including emojis
- **Input Validation**: GORM handles SQL injection prevention

## 🧪 Sample Data

The seed file includes:
- 4 sample URLs (example.com, github.com, stackoverflow.com, w3.org)
- 3 completed crawl sessions with realistic data
- 11 sample links with various status codes
- Mix of internal/external and accessible/broken links

This data allows immediate testing of:
- Dashboard functionality
- Charts and visualizations
- Broken link detection
- Filter and search features 