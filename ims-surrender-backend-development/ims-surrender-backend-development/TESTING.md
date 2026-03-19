# Surrender Service Testing Guide

## Overview
This document provides comprehensive testing guidelines for the Surrender Service.

## Test Structure

```
surrender-service/
├── core/domain/
│   └── domain_test.go              # Domain model tests
├── repo/postgres/
│   └── surrender_request_test.go   # Repository tests
├── handler/
│   └── voluntary_surrender_test.go # Handler tests
├── temporal/
│   └── workflows/workflow_test.go  # Workflow tests
└── scripts/
    └── run_tests.sh                # Test runner script
```

## Running Tests

### All Tests
```bash
go test ./...
```

### With Coverage
```bash
go test -cover ./...
```

### Specific Package
```bash
go test -v ./handler/...
```

### With Race Detector
```bash
go test -race ./...
```

### Using Test Script
```bash
chmod +x scripts/run_tests.sh
./scripts/run_tests.sh
```

### With Benchmarks
```bash
./scripts/run_tests.sh --bench
```

## Test Categories

### 1. Domain Model Tests
**Location:** `core/domain/domain_test.go`

Tests for:
- Enum validation
- Model validation
- Field constraints
- Metadata handling
- JSON marshaling

### 2. Repository Tests
**Location:** `repo/postgres/*_test.go`

Tests for:
- CRUD operations
- Batch operations
- Query filtering
- Transaction handling
- Error scenarios

**Requirements:**
- Test database connection
- Database schema setup
- Test data fixtures

**Setup:**
```bash
# Create test database
createdb surrender_test

# Run migrations
psql surrender_test < migrations/001_init_schema.sql
```

### 3. Handler Tests
**Location:** `handler/*_test.go`

Tests for:
- Request validation
- Business logic
- Response formatting
- Error handling
- Mock dependencies

**Mocking:**
- Uses `testify/mock` for dependencies
- Mocks repositories and external services
- Tests HTTP request/response cycle

### 4. Workflow Tests
**Location:** `temporal/workflows/workflow_test.go`

Tests for:
- Workflow execution paths
- Signal handling
- Activity mocking
- Timeout scenarios
- Error recovery

**Using Temporal Test Suite:**
```go
testSuite := &testsuite.WorkflowTestSuite{}
env := testSuite.NewTestWorkflowEnvironment()
```

## Test Data

### Fixtures
Create test data fixtures in `testdata/` directory:
```
testdata/
├── policies.json
├── surrender_requests.json
└── documents.json
```

### Test Helpers
Common test utilities in each package:
```go
func setupTestDB(t *testing.T) *pgxpool.Pool
func createTestSurrenderRequest(t *testing.T) *domain.PolicySurrenderRequest
func mockPolicyService() *MockPolicyService
```

## Coverage Goals

| Component | Target Coverage |
|-----------|----------------|
| Domain Models | 90%+ |
| Repositories | 80%+ |
| Handlers | 85%+ |
| Workflows | 75%+ |
| Overall | 80%+ |

## Viewing Coverage

### HTML Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Terminal Summary
```bash
go test -cover ./...
```

### Per-Function Coverage
```bash
go tool cover -func=coverage.out
```

## Best Practices

### 1. Table-Driven Tests
```go
tests := []struct {
    name    string
    input   InputType
    want    OutputType
    wantErr bool
}{
    {"valid case", validInput, expectedOutput, false},
    {"error case", invalidInput, nil, true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test implementation
    })
}
```

### 2. Test Naming
- `TestFunctionName` for unit tests
- `TestFunctionName_Scenario` for specific scenarios
- `BenchmarkFunctionName` for benchmarks

### 3. Assertions
Use `testify/assert` and `testify/require`:
```go
require.NoError(t, err)  // Stops test on failure
assert.Equal(t, expected, actual)  // Continues on failure
```

### 4. Cleanup
```go
t.Cleanup(func() {
    // Cleanup code
})
```

### 5. Skip Long Tests
```go
if testing.Short() {
    t.Skip("Skipping long-running test")
}
```

## Integration Tests

### Database Integration
```bash
# Run only integration tests
go test -tags=integration ./repo/...
```

### Temporal Integration
```bash
# Requires Temporal server running
docker-compose up -d temporal
go test ./temporal/...
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.25
      - name: Run tests
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
```

## Troubleshooting

### Test Database Connection Failed
```bash
# Check PostgreSQL is running
pg_isready

# Verify connection string
echo $DATABASE_URL
```

### Mock Expectations Not Met
```go
// Add at end of test
mockRepo.AssertExpectations(t)
```

### Workflow Test Timeout
```go
// Increase test timeout
env := testSuite.NewTestWorkflowEnvironment()
env.SetTestTimeout(time.Minute * 5)
```

## Performance Testing

### Benchmarks
```bash
go test -bench=. -benchmem ./...
```

### Example Benchmark
```go
func BenchmarkCalculateSurrenderValue(b *testing.B) {
    for i := 0; i < b.N; i++ {
        calculateSurrenderValue(testPolicy)
    }
}
```

### Load Testing
Use `k6` or `vegeta` for API load testing:
```bash
k6 run scripts/load_test.js
```

## Security Testing

### SQL Injection
Test repository methods with malicious inputs

### Input Validation
Test all handler endpoints with invalid data

### Authentication/Authorization
Test access control (when implemented)

## Continuous Improvement

1. Monitor coverage trends
2. Add tests for bug fixes
3. Review and update test data
4. Refactor duplicate test code
5. Document complex test scenarios

## Resources

- [Go Testing](https://golang.org/pkg/testing/)
- [Testify](https://github.com/stretchr/testify)
- [Temporal Testing](https://docs.temporal.io/docs/go/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
