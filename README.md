# Web Crawler Application

A comprehensive web crawling application that analyzes websites and provides detailed insights about their structure, links, and SEO metrics.

## 🚀 Features

### Backend Data Collection
- **HTML Version Detection** - Automatically detects HTML5, XHTML, etc.
- **Page Title Extraction** - Retrieves and displays page titles
- **Heading Analysis** - Counts heading tags by level (H1, H2, H3, H4, H5, H6)
- **Link Analysis** - Categorizes internal vs external links
- **Broken Link Detection** - Identifies inaccessible links (4xx/5xx status codes)
- **Login Form Detection** - Detects presence of authentication forms

### Frontend Dashboard
- **URL Management** - Add, view, and delete website URLs
- **Real-time Status Updates** - Live crawl progress monitoring
- **Interactive Data Table** - Paginated, sortable table with filters and global search
- **Detailed Analytics** - Charts for link distribution and SEO insights
- **Bulk Operations** - Re-run analysis or delete multiple URLs at once
- **Responsive Design** - Modern UI optimized for all devices

## ✅ Project Status

**What's Included:**
- ✅ **Complete Backend** - Full Go API with database, migrations, and business logic
- ✅ **Complete Frontend** - React dashboard with modern UI and real-time updates
- ✅ **Docker Setup** - Production-ready containerization
- ✅ **Database** - MySQL with proper schema and migrations
- ✅ **Tests** - Comprehensive test suite for both backend and frontend
- ✅ **Documentation** - Detailed README with setup instructions
- ✅ **Error Handling** - Robust error handling and user feedback
- ✅ **Security** - JWT authentication and input validation

**Ready to Use:**
- 🚀 **One-command setup**: `docker compose up -d`
- 🧪 **Full test coverage**: Backend and frontend tests passing
- 📱 **Responsive design**: Works on desktop, tablet, and mobile
- 🔧 **Development ready**: Hot reload, debugging, and local development setup

## 🛠 Technology Stack

### Frontend
- **React 18** with TypeScript
- **Vite** for fast development and building
- **React Router** for navigation
- **React Query** for data fetching and caching
- **Tailwind CSS** for styling
- **Recharts** for data visualization
- **React Hot Toast** for notifications

### Backend
- **Go 1.21** with Gin framework
- **GORM** for database operations
- **MySQL 8.0** database
- **golang-migrate** for database migrations
- **Clean Architecture** with layered design

### DevOps
- **Docker & Docker Compose** for containerization
- **Nginx** for frontend serving and API proxying
- **Multi-stage builds** for optimized production images

## 📋 Prerequisites

- Docker Desktop installed and running
- Git for cloning the repository
- Modern web browser

## 🧪 Testing

### Backend Tests
```bash
cd backend
make test          # Run all tests
make test-verbose  # Run tests with verbose output
```

**Test Coverage:**
- ✅ **Models** - Database models and relationships
- ✅ **Services** - Business logic (URL, Crawler, Auth services)
- ✅ **Handlers** - HTTP request handling
- ✅ **Middleware** - Authentication and error handling

### Frontend Tests
```bash
cd my-vite-project
npm test           # Run all tests
npm test -- --watch  # Run tests in watch mode
```

**Test Coverage:**
- ✅ **Components** - Dashboard, AddUrl, and utility components
- ✅ **User Interactions** - Form validation, button clicks, navigation
- ✅ **API Integration** - Mock service calls and error handling

### Running All Tests
```bash
# Backend tests
cd backend && make test

# Frontend tests  
cd my-vite-project && npm test

# Or run both from root
cd backend && make test && cd ../my-vite-project && npm test
```

## ⚡ Quick Start

### 1. Clone the Repository
```bash
git clone <your-repo-url>
cd web-dev
```

### 2. Installation 

#### Option A: Docker Compose (Recommended)
The easiest way to run the entire application:
```bash
# Start all services
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f
```

#### Option B: Local Development
For development with hot reload:

1. **Backend Setup**:
```bash
cd backend
go mod download
go run main.go
```

2. **Frontend Setup**:
```bash
cd my-vite-project
npm install
npm run dev
```

3. **Database Setup**:
```bash
# Using Docker for MySQL only
docker run --name mysql-webcrawler \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=webcrawler \
  -p 3306:3306 \
  -d mysql:8.0

# Or use the full docker-compose for database
docker compose up mysql -d
```

### 3. Access the Application
Open your browser and navigate to: **http://localhost:3000**

### 4. Run Tests (Optional)
```bash
# Backend tests
cd backend && make test

# Frontend tests
cd my-vite-project && npm test
```

## 📖 Usage Guide

### Adding a Website for Analysis
1. Click the **"+ Add URL"** button in the top navigation
2. Enter the website URL (e.g., `https://example.com`)
3. Click **"Add URL"** to start the crawling process

### Viewing Analysis Results
1. From the dashboard, locate your URL in the table
2. Click the **"View"** button to see detailed analysis
3. Explore charts showing:
   - Internal vs External link distribution
   - Link accessibility status
   - Heading structure analysis
   - SEO score metrics

### Managing URLs
- **Delete**: Use the "Delete" button for individual URLs
- **Bulk Operations**: Select multiple URLs using checkboxes and use bulk actions
- **Re-run Analysis**: Restart crawling for updated results

## 🔗 API Documentation

### Base URL
```