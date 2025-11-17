# MySQL Support Design

## Current State Analysis

The application currently uses SQLite exclusively with the following characteristics:
- Pure Go SQLite implementation (`github.com/glebarez/sqlite`)
- Single database file storage (`runtime/data/nano.db`)
- WAL mode for concurrency
- GORM ORM for database operations
- Automatic schema migration on startup
- Simple configuration with only `dbname` parameter

## Proposed Architecture

### Database Abstraction Layer
Create a database abstraction that can handle multiple database types:
1. Database type detection from configuration
2. Database-specific connection string generation
3. Unified database initialization interface
4. Shared GORM configuration and migration logic

### Configuration Changes
Extend the existing `Database` configuration structure:
- Add `Type` field to specify database type (sqlite, mysql)
- Maintain backward compatibility with existing SQLite configuration
- Add MySQL-specific fields (User, Password, Host, Port already exist)

### Connection Logic
Implement database-type-aware connection logic:
1. Parse configuration to determine database type
2. Generate appropriate connection string
3. Initialize GORM with the correct dialect
4. Apply database-specific optimizations

### Database Type Detection
The database type will be determined by:
1. New `type` configuration field (if present)
2. Database name pattern (`.db` extension implies SQLite)
3. Default to SQLite for backward compatibility

### Migration Strategy
Ensure schema migrations work across both databases:
1. Use GORM's database-agnostic migration features
2. Test migrations on both SQLite and MySQL
3. Handle any database-specific differences in model definitions

## Technical Considerations

### Dependencies
Add MySQL driver dependency:
```go
gorm.io/driver/mysql
```

### Connection Management
Implement proper connection pooling for MySQL:
- Configure appropriate pool sizes
- Set connection timeouts
- Handle connection lifecycle properly

### Error Handling
Ensure database-specific errors are handled appropriately:
- Connection failures
- Authentication errors
- Database-specific constraint violations

### Performance Optimization
Apply database-specific optimizations:
- SQLite: WAL mode, connection pooling
- MySQL: Connection pooling, query optimization

## Implementation Steps
1. Add MySQL driver dependency
2. Extend database configuration structure
3. Implement database type detection logic
4. Create database-agnostic initialization function
5. Update models for cross-database compatibility
6. Test with both SQLite and MySQL
7. Update documentation