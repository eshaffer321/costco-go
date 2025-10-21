# Developer Guide for Claude Code

This document contains important information for AI assistants (like Claude) working on this codebase.

## Project Overview

This is a Go client library and CLI for accessing Costco order history and receipt data via their GraphQL API. The library handles OAuth2 authentication, automatic token refresh, and provides a clean Go interface for fetching orders and receipts.

## Project Structure

```
costco-go/
├── cmd/costco-cli/           # CLI application
│   └── main.go               # CLI entry point
├── pkg/costco/               # Core library package
│   ├── client.go             # Main client implementation
│   ├── auth.go               # Authentication logic
│   ├── orders.go             # Order-related operations
│   ├── receipts.go           # Receipt-related operations
│   ├── constants.go          # API constants and configuration
│   └── *_test.go             # Test files
├── CHANGELOG.md              # Version history
├── README.md                 # User documentation
└── go.mod                    # Go module definition
```

## IMPORTANT: Versioning Process

### ⚠️ CRITICAL RULE: Every PR Must Include a Version Bump

**Every pull request that changes code MUST include a version bump.** This is non-negotiable.

- **Bug fixes** → Bump PATCH version (e.g., 0.1.0 → 0.1.1)
- **New features** → Bump MINOR version (e.g., 0.1.0 → 0.2.0)
- **Breaking changes** → Bump MAJOR version (e.g., 0.1.0 → 1.0.0)
- **Documentation-only changes** → No version bump needed

If you're making a code change and don't bump the version, the PR should be rejected.

### Versioning Steps

This library follows [Semantic Versioning](https://semver.org/). When releasing a new version, you **MUST** follow these steps in order:

### Step 1: Update the Version Constant
Edit `pkg/costco/constants.go` and update the `Version` constant:

```go
// Library Version
const (
	Version = "X.Y.Z"  // Update this line
)
```

**Version number rules:**
- **MAJOR** (X.0.0): Breaking API changes that are not backwards compatible
- **MINOR** (0.Y.0): New features that are backwards compatible
- **PATCH** (0.0.Z): Bug fixes that are backwards compatible

### Step 2: Update CHANGELOG.md
Add a new section at the top of `CHANGELOG.md` following this format:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features go here

### Changed
- Changes to existing functionality go here

### Deprecated
- Soon-to-be removed features go here

### Removed
- Removed features go here

### Fixed
- Bug fixes go here

### Security
- Security-related changes go here

[X.Y.Z]: https://github.com/costco-go/compare/vX.Y.Z
```

**Only include sections that have actual changes.** For example, if you only fixed bugs, just include the "Fixed" section.

### Step 3: Update README Badge
Update the version badge in `README.md`:

```markdown
[![Version](https://img.shields.io/badge/version-X.Y.Z-blue.svg)](https://github.com/costco-go/releases/tag/vX.Y.Z)
```

### Step 4: Create and Push Git Tag
Create an annotated git tag and push it:

```bash
git tag vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

### Versioning Checklist

Before releasing a new version, ensure:

- [ ] `pkg/costco/constants.go` Version constant is updated
- [ ] `CHANGELOG.md` has a new section documenting all changes
- [ ] `README.md` version badge is updated
- [ ] All tests pass (`go test ./pkg/costco -v`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] Git tag is created with the version (e.g., `v0.2.0`)
- [ ] Git tag is pushed to remote (`git push origin vX.Y.Z`)

## Pull Request Workflow

### Before Creating a PR

Every PR that includes code changes MUST:

1. **Determine the version bump type:**
   - Bug fix only? → PATCH (0.1.0 → 0.1.1)
   - New feature? → MINOR (0.1.0 → 0.2.0)
   - Breaking change? → MAJOR (0.1.0 → 1.0.0)

2. **Update all three versioning files:**
   - `pkg/costco/constants.go` - Update the `Version` constant
   - `CHANGELOG.md` - Add a new section with your changes
   - `README.md` - Update the version badge

3. **Run all tests:**
   ```bash
   go test ./pkg/costco -v
   go fmt ./...
   ```

4. **Commit everything together:**
   ```bash
   git add .
   git commit -m "Add feature X - bump version to 0.2.0"
   ```

5. **After the PR is merged, create the git tag:**
   ```bash
   git tag v0.2.0 -m "Release v0.2.0"
   git push origin v0.2.0
   ```

### Example PR Description Template

```markdown
## Changes
- Added new method to fetch membership details
- Fixed bug in receipt parsing

## Version Bump
- Type: MINOR (0.1.0 → 0.2.0)
- Reason: New feature added

## Checklist
- [x] Version constant updated in constants.go
- [x] CHANGELOG.md updated
- [x] README.md badge updated
- [x] All tests passing
- [x] Code formatted with gofmt
```

## Testing

Always run tests before creating a PR:

```bash
# Run all tests
go test ./pkg/costco -v

# Run tests with coverage
go test ./pkg/costco -v -cover

# Run specific test
go test ./pkg/costco -v -run TestClientGetOnlineOrders
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and testable
- Use structured logging with `log/slog`
- Add comments for exported functions and types

## Authentication Details

The client uses Azure AD B2C OAuth2/OIDC flow:
- Initial authentication requires email/password
- Returns both access token and refresh token
- Automatically refreshes tokens before expiry
- Thread-safe token management with mutex locks

## API Constants

All API endpoints, client IDs, and configuration values are in `pkg/costco/constants.go`. These are public values used by the Costco website and mobile app.

**IMPORTANT:** Never commit actual user credentials. The constants file only contains public OAuth2 client identifiers.

## Logging

- Default: Silent mode using `io.Discard`
- Optional: Inject custom `*slog.Logger` via `Config.Logger`
- All logs include `client=costco` attribute for filtering
- Use appropriate log levels: Debug, Info, Warn, Error

## Common Tasks

### Adding a New API Method

1. Define the GraphQL query in the appropriate file (orders.go, receipts.go)
2. Create response struct types
3. Implement the client method
4. Add comprehensive tests with mocked HTTP responses
5. Update README.md with usage example
6. Document in CHANGELOG.md under "Added"

### Fixing a Bug

1. Write a failing test that reproduces the bug
2. Fix the bug
3. Verify the test passes
4. Document in CHANGELOG.md under "Fixed"
5. Consider if this requires a patch version bump

### Breaking Changes

If you need to make a breaking change:
1. Consider if there's a backwards-compatible way
2. If not, clearly document the breaking change in CHANGELOG.md
3. Bump the MAJOR version number
4. Update README.md with migration guide if needed

## Dependencies

Keep dependencies minimal:
- `github.com/golang-jwt/jwt/v5` - JWT token parsing
- `github.com/stretchr/testify` - Testing assertions (dev only)

Before adding a new dependency, consider:
- Is it really needed?
- Is it actively maintained?
- Does it have a reasonable license?
- Does it increase the attack surface?

## Security Considerations

- Never log passwords or tokens (even at Debug level)
- Use HTTPS for all API calls
- Validate JWT tokens properly
- Handle token expiry gracefully
- Don't store credentials in code or config files

## Release Workflow Example

Here's a complete example of releasing version 0.2.0 with a new feature:

```bash
# 1. Make your changes and test them
go test ./pkg/costco -v

# 2. Update pkg/costco/constants.go
# Change: Version = "0.1.0"
# To:     Version = "0.2.0"

# 3. Update CHANGELOG.md
# Add new section:
## [0.2.0] - 2025-10-20

### Added
- New method to fetch membership details

# 4. Update README.md version badge
# Change: version-0.1.0
# To:     version-0.2.0

# 5. Commit your changes
git add .
git commit -m "Release v0.2.0"

# 6. Create and push tag
git tag v0.2.0 -m "Release v0.2.0 - Add membership details API"
git push origin main
git push origin v0.2.0
```

Now users can install with:
```bash
go get github.com/costco-go/pkg/costco@v0.2.0
```

## Questions or Issues?

If you're unsure about versioning or any other aspect of the codebase, ask the maintainer before making changes.
