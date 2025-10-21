# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2025-10-19

### Fixed
- Improved receipt fetching logging to eliminate misleading ERROR logs during normal operation
- Optimized `GetReceipts()` fallback logic to try object format first (the format Costco API currently returns), reducing unnecessary failed attempts
- Added clear diagnostic logging with emojis to identify if array fallback is ever needed
- Reduced log noise by changing generic GraphQL decode errors to DEBUG level

[0.1.1]: https://github.com/eshaffer321/costco-go/compare/v0.1.0...v0.1.1

## [0.1.0] - 2025-10-19

### Added
- Initial release of Costco Go client library
- Authentication using Azure AD B2C OAuth2/OIDC flow
- Order history retrieval via GraphQL API
- Support for refresh tokens
- Structured logging with slog (silent by default using io.Discard)
- CLI tool for command-line usage
- Configurable warehouse selection
- Pagination support for order history

### Fixed
- Azure AD B2C authentication flow
- GraphQL array response handling
- Linting and formatting issues
- Test failures with structured logging

[0.1.0]: https://github.com/costco-go/compare/v0.1.0
