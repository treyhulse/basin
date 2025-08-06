# Testing Documentation

This directory contains testing utilities and documentation for the Basin API.

## Files

### PowerShell Test Scripts
- **[field-api-test.ps1](field-api-test.ps1)** - PowerShell script for testing field API operations

## Running Tests

### Automated Tests
The main test suite is located in the project root and can be run with:

```bash
# Clean output
make test

# Verbose output with all test details
make test-verbose

# With coverage report
make test-coverage

# Unit tests only
make test-unit

# Integration tests only
make test-integration
```

### Manual Testing Scripts
The PowerShell scripts in this directory can be used for manual testing and debugging:

```powershell
# Run field API tests
.\documentation\testing\field-api-test.ps1
```

## Test Structure

The automated test suite covers:

### Unit Tests
- **Authentication** (`internal/api/auth_test.go`)
  - Login validation
  - Token format validation
  - Security headers
  - Error handling

- **Items API** (`internal/api/items_test.go`)
  - CRUD operations
  - Input validation
  - Security checks
  - UUID validation

- **RBAC Policies** (`internal/rbac/policies_test.go`)
  - Permission checking
  - Field filtering
  - Row-level security

- **Configuration** (`internal/config/env_test.go`)
  - Environment variable validation
  - Default values
  - Connection string building

### Integration Tests
- **Full API Flow** (`test/integration_test.go`)
  - Complete authentication flow
  - End-to-end API operations
  - CORS headers
  - Health checks

## Test Coverage

Current test coverage includes:
- ✅ Authentication endpoints
- ✅ Dynamic CRUD operations
- ✅ Role-based access control
- ✅ Input validation and security
- ✅ Environment configuration
- ✅ Integration flows

## Writing New Tests

When adding new features, ensure you:
1. Add unit tests for individual functions
2. Add integration tests for complete flows
3. Test both success and error scenarios
4. Include security and validation tests
5. Update documentation as needed

### Test Naming Convention
- Test files: `*_test.go`
- Test functions: `TestFunctionName`
- Test cases: Use descriptive names in table-driven tests

### Example Test Structure
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name           string
        input          string
        expectedOutput string
        expectError    bool
    }{
        {
            name:           "Valid input",
            input:          "valid",
            expectedOutput: "expected",
            expectError:    false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```