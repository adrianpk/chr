# Go Conventions

This project uses Go's standard library by default.

## Testing

- Use Go's standard `testing` package.
- For bug fixes, first reproduce or observe the defect with a failing test when
  feasible.
- Prefer table-driven tests for input/output behavior.
- Place tests beside the package they validate.
- Cover expected failures, edge cases, and package boundary contracts.
- Use property-based, fuzz, race, or benchmark tests only when the use case
  justifies them.
