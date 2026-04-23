# Groven Backend Developer Skill

## Purpose
This skill provides specialized assistance for developing the Groven backend API in Go.

## Available Commands

### `/backend-init`
Initialize the Go backend project structure and dependencies.

### `/backend-model`
Create a new data model and corresponding database operations.

### `/backend-handler`
Create a new HTTP handler for a specific endpoint.

### `/backend-service`
Create a business logic service for a specific domain.

### `/backend-repo`
Create a new repository implementation (GORM or sqlx).

### `/backend-migration`
Create a new database migration file.

## Usage Examples

```
/backend-init                    # Initialize project
/backend-model agent            # Create Agent model
/backend-handler agents         # Create agents endpoint handlers
/backend-service agents         # Create agents business logic
/backend-repo agents gorm       # Create GORM repository for agents
/backend-migration add_agents   # Create migration for agents table
```

## Dependencies

This skill assumes:
- Go 1.21+ installed
- PostgreSQL with TimescaleDB
- Redis
- Docker and Docker Compose

## Project Structure

The skill works with the following backend structure:

```
api/
├── cmd/server/
│   └── main.go              # Application entry point
├── internal/
│   ├── handler/             # HTTP handlers
│   ├── service/             # Business logic
│   ├── model/               # Data models
│   ├── repository/          # Data access
│   └── middleware/          # Middleware
└── pkg/                     # Public packages
```

## Key Principles

1. **Clean Architecture**: Separation of concerns between layers
2. **Dependency Injection**: Use interfaces for testability
3. **Error Handling**: Consistent error responses and logging
4. **Performance**: Use sqlx/raw SQL for performance-critical paths
5. **Validation**: Input validation at handler level
6. **Security**: Proper authentication and authorization

## Common Patterns

### Handler Pattern
```go
func (h *Handler) CreateAgent(w http.ResponseWriter, r *http.Request) {
    // Parse request
    // Validate input
    // Call service
    // Return response
}
```

### Service Pattern
```go
func (s *Service) CreateAgent(ctx context.Context, req *CreateAgentRequest) (*Agent, error) {
    // Business logic
    // Call repository
    // Return result
}
```

### Repository Pattern
```go
type AgentRepository interface {
    Create(ctx context.Context, agent *Agent) error
    GetByID(ctx context.Context, id string) (*Agent, error)
    Update(ctx context.Context, agent *Agent) error
    Delete(ctx context.Context, id string) error
}
```

## Database Operations

### GORM Operations
- Simple CRUD operations
- Relationship handling
- Automatic migrations

### sqlx Operations
- Complex queries
- Performance-critical paths
- Raw SQL for optimization

## Testing

### Unit Tests
- Test handlers with mock services
- Test services with mock repositories
- Test business logic in isolation

### Integration Tests
- Test database operations
- Test API endpoints
- Test authentication flows

## Monitoring

### Health Checks
- Database connectivity
- Redis connectivity
- Application status

### Metrics
- Request latency
- Error rates
- Database query performance

## Security

### Authentication
- API key authentication
- JWT tokens for protected endpoints
- Rate limiting

### Authorization
- Role-based access control
- Resource ownership checks
- Audit logging

## Environment Variables

### Required
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `GROVEN_API_KEY` - API key for authentication

### Optional
- `LOG_LEVEL` - Logging level (debug, info, warn, error)
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment (development, staging, production)