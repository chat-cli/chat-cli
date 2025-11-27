# Test Driven Development

When developing features or fixing bugs in this project, follow Test Driven Development (TDD) practices:

1. **Write tests first** - Before implementing any new functionality, write the test that defines the expected behavior
2. **Run the test and watch it fail** - Verify the test fails for the right reason
3. **Write minimal code** - Implement just enough code to make the test pass
4. **Run tests and verify they pass** - Ensure the new test passes along with all existing tests
5. **Refactor** - Clean up the code while keeping tests green

## Running Tests

Use the following commands to run tests:

```bash
# Run all tests
make test

# Run tests in short mode (skip integration tests)
make test-short

# Run tests with coverage
make test-coverage
```

## Test Guidelines

- Keep tests focused and test one thing at a time
- Use descriptive test names that explain what is being tested
- Follow Go testing conventions and use the `testing` package
- Place tests in `*_test.go` files alongside the code they test
- Aim for high test coverage, especially for critical business logic
