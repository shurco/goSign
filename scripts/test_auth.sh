#!/usr/bin/env bash

# Authentication Tests Runner
# This script runs all authentication-related tests with various options

set -e

ROOT_PATH="$(git rev-parse --show-toplevel)"
cd "${ROOT_PATH}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Print header
print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# Print success message
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Print error message
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Print info message
print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Check if tests exist
check_tests() {
    if [ ! -f "internal/handlers/public/auth_test.go" ]; then
        print_error "Auth tests not found!"
        exit 1
    fi
    print_success "Test files found"
}

# Run unit tests
run_unit_tests() {
    print_header "Running Unit Tests"
    
    print_info "Running auth unit tests..."
    if go test ./internal/handlers/public/... -v -timeout 60s; then
        print_success "Unit tests passed"
    else
        print_error "Unit tests failed"
        exit 1
    fi
}

# Run tests with coverage
run_coverage() {
    print_header "Running Tests with Coverage"
    
    print_info "Generating coverage report..."
    if go test ./internal/handlers/public/... -cover -coverprofile=coverage.out -covermode=atomic; then
        print_success "Coverage report generated"
        
        # Display coverage summary
        echo ""
        go tool cover -func=coverage.out | tail -n 1
        
        # Generate HTML report
        print_info "Generating HTML coverage report..."
        go tool cover -html=coverage.out -o coverage.html
        print_success "HTML report saved to coverage.html"
    else
        print_error "Coverage generation failed"
        exit 1
    fi
}

# Run benchmarks
run_benchmarks() {
    print_header "Running Benchmark Tests"
    
    print_info "Running password hashing benchmarks..."
    go test ./internal/handlers/public/ -bench=BenchmarkPassword -benchmem -run=^$ | grep -E "Benchmark|ns/op|allocs/op"
    print_success "Benchmarks completed"
}

# Run specific test
run_specific_test() {
    local test_name=$1
    print_header "Running Specific Test: ${test_name}"
    
    if go test ./internal/handlers/public/ -run "${test_name}" -v; then
        print_success "Test passed: ${test_name}"
    else
        print_error "Test failed: ${test_name}"
        exit 1
    fi
}

# Run race detector
run_race_tests() {
    print_header "Running Tests with Race Detector"
    
    print_info "Checking for race conditions..."
    if go test ./internal/handlers/public/... -race -short; then
        print_success "No race conditions detected"
    else
        print_error "Race conditions found!"
        exit 1
    fi
}

# Clean up
cleanup() {
    print_info "Cleaning up test artifacts..."
    rm -f coverage.out coverage.html
    print_success "Cleanup complete"
}

# Display help
show_help() {
    cat << EOF
Authentication Test Runner

Usage: $0 [OPTION]

OPTIONS:
    unit        Run unit tests only
    coverage    Run tests with coverage report
    bench       Run benchmark tests
    race        Run tests with race detector
    specific    Run specific test (provide test name)
    all         Run all tests (default)
    clean       Clean up test artifacts
    help        Show this help message

EXAMPLES:
    $0 unit                         # Run unit tests
    $0 coverage                     # Generate coverage report
    $0 bench                        # Run benchmarks
    $0 specific TestSignUpFlow      # Run specific test
    $0 all                          # Run all tests

TEST CATEGORIES:
    - Sign Up Flow Tests
    - Sign In Flow Tests
    - Password Management Tests
    - Email Verification Tests
    - Token Management Tests
    - Security Feature Tests
    - User Role Tests

EOF
}

# Main execution
main() {
    local command=${1:-all}
    
    case "$command" in
        unit)
            check_tests
            run_unit_tests
            ;;
        coverage)
            check_tests
            run_coverage
            ;;
        bench)
            check_tests
            run_benchmarks
            ;;
        race)
            check_tests
            run_race_tests
            ;;
        specific)
            if [ -z "$2" ]; then
                print_error "Please provide test name"
                echo "Example: $0 specific TestSignUpFlow"
                exit 1
            fi
            check_tests
            run_specific_test "$2"
            ;;
        all)
            check_tests
            run_unit_tests
            echo ""
            run_coverage
            echo ""
            run_benchmarks
            print_header "Test Summary"
            print_success "All tests completed successfully!"
            ;;
        clean)
            cleanup
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"

