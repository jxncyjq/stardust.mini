# Skills Package

The skills package provides a complete implementation for managing skills with database persistence and HTTP API endpoints.

## Features

- **Database Entity**: Skill model with GORM annotations for MySQL/PostgreSQL
- **DAO Pattern**: Data Access Object for CRUD operations
- **HTTP Handlers**: RESTful API endpoints using the framework's handler pattern
- **Filtering**: Support for filtering by category and level
- **Testing**: Comprehensive unit tests with mocks

## Database Schema

The `Skill` entity has the following fields:

```go
type Skill struct {
    ID          int64     // Primary key, auto-increment
    Name        string    // Skill name (required)
    Description string    // Detailed description
    Level       string    // Skill level (e.g., Beginner, Intermediate, Advanced)
    Category    string    // Skill category (e.g., Programming, DevOps, Design)
    CreatedAt   time.Time // Auto-populated on creation
    UpdatedAt   time.Time // Auto-updated on modification
}
```

## Usage

### 1. Setup Database Connection

```go
import (
    "github.com/jxncyjq/stardust.mini/databases"
    "github.com/jxncyjq/stardust.mini/skills"
)

// Initialize database
dbConfig := []byte(`{"dsn": "user:pass@tcp(localhost:3306)/dbname"}`)
db, err := databases.NewMySQLDB(dbConfig)
if err != nil {
    panic(err)
}

dao := databases.NewBaseDao(db.DB())
```

### 2. Create and Register HTTP Endpoints

```go
import (
    httpServer "github.com/jxncyjq/stardust.mini/http_server"
)

// Create HTTP server
httpConfig := []byte(`{"address":"127.0.0.1","port":8080,"path":"/api"}`)
srv, err := httpServer.NewHttpServer(httpConfig)
if err != nil {
    panic(err)
}

// Initialize skill handler
skillHandler := skills.NewSkillHandler(dao)

// Register list endpoint
srv.Get("skills", "", skillHandler.ListSkillsHandler())

// Start server
if err := srv.Startup(); err != nil {
    panic(err)
}
```

### 3. API Endpoints

Once registered, the following endpoints are available:

- **GET /api/skills** - List all skills
  ```bash
  curl http://localhost:8080/api/skills
  ```

- **GET /api/skills?category=Programming** - Filter by category
  ```bash
  curl http://localhost:8080/api/skills?category=Programming
  ```

- **GET /api/skills?level=Advanced** - Filter by level
  ```bash
  curl http://localhost:8080/api/skills?level=Advanced
  ```

### 4. Response Format

```json
{
  "skills": [
    {
      "id": 1,
      "name": "Go Programming",
      "description": "Backend development with Go",
      "level": "Advanced",
      "category": "Programming",
      "created_at": "2026-03-03T01:00:00Z",
      "updated_at": "2026-03-03T01:00:00Z"
    }
  ],
  "total": 1
}
```

## Using DAO Directly

### Create Skills

```go
skillDao := skills.NewSkillDao(dao)

skill := &skills.Skill{
    Name:        "Go Programming",
    Description: "Backend development with Go",
    Level:       "Advanced",
    Category:    "Programming",
}

entity := skills.NewSkillEntity(dao, skill)
rows, err := entity.Create()
```

### Query Skills

```go
// List all skills
allSkills, err := skillDao.ListAll()

// Filter by category
programmingSkills, err := skillDao.ListByCategory("Programming")

// Filter by level
advancedSkills, err := skillDao.ListByLevel("Advanced")

// Find by ID
skill, err := skillDao.FindByID(1)
```

## Database Migration

To create the skills table, use GORM's auto-migration:

```go
dao.Migrations([]interface{}{&skills.Skill{}})
```

This will create the following table:

```sql
CREATE TABLE skills (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    level VARCHAR(50),
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

## Testing

Run the tests:

```bash
go test ./skills -v
```

The package includes comprehensive tests for:
- Listing all skills
- Filtering by category
- Filtering by level
- Entity methods (TableName, PrimaryKey)
