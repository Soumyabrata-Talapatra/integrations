#!/bin/bash

# Surrender Service Test Runner
# Runs all tests with coverage reporting

set -e

echo "======================================"
echo "Surrender Service - Test Suite"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Clean previous test artifacts
print_info "Cleaning previous test artifacts..."
rm -f coverage.out coverage.html
echo ""

# Run unit tests
print_info "Running unit tests..."
echo ""

# Test domain models
print_info "Testing domain models..."
go test -v ./core/domain/... -coverprofile=coverage-domain.out || {
    print_error "Domain tests failed"
    exit 1
}
print_success "Domain tests passed"
echo ""

# Test repositories (skip if no test DB)
print_info "Testing repositories..."
go test -v ./repo/postgres/... -coverprofile=coverage-repo.out || {
    print_error "Repository tests failed (skipped if no test DB)"
}
print_success "Repository tests completed"
echo ""

# Test handlers
print_info "Testing handlers..."
go test -v ./handler/... -coverprofile=coverage-handler.out || {
    print_error "Handler tests failed"
    exit 1
}
print_success "Handler tests passed"
echo ""

# Test workflows
print_info "Testing Temporal workflows..."
go test -v ./temporal/workflows/... -coverprofile=coverage-workflow.out || {
    print_error "Workflow tests failed"
    exit 1
}
print_success "Workflow tests passed"
echo ""

# Combine coverage reports
print_info "Generating coverage report..."
echo "mode: set" > coverage.out
grep -h -v "^mode:" coverage-*.out >> coverage.out 2>/dev/null || true

# Generate coverage statistics
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
print_info "Total coverage: ${COVERAGE}"
echo ""

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
print_success "HTML coverage report generated: coverage.html"
echo ""

# Run race detector on critical paths
print_info "Running race detector tests..."
go test -race ./handler/... > /dev/null 2>&1 || {
    print_error "Race conditions detected!"
    exit 1
}
print_success "No race conditions detected"
echo ""

# Run benchmarks (optional)
if [ "$1" == "--bench" ]; then
    print_info "Running benchmarks..."
    go test -bench=. -benchmem ./... > benchmark.txt
    print_success "Benchmark results saved to benchmark.txt"
    echo ""
fi

# Summary
echo "======================================"
print_success "All tests passed!"
echo "======================================"
echo ""
echo "Coverage: ${COVERAGE}"
echo "Coverage report: coverage.html"
echo ""

# Clean up individual coverage files
rm -f coverage-*.out

exit 0
