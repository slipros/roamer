# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

### Documentation

## [1.16.0] - 2025-10-04

### Added
- Preserve body functionality to read request body multiple times ([#545017d](https://github.com/slipros/roamer/commit/545017d))
- `NewParseWithPool` function for parsing with custom sync.Pool support ([#5bab084](https://github.com/slipros/roamer/commit/5bab084))
- Extensive test coverage improvements ([#fc91496](https://github.com/slipros/roamer/commit/fc91496), [#729c837](https://github.com/slipros/roamer/commit/729c837))

### Changed
- Performance optimizations in core parsing logic ([#f2634a8](https://github.com/slipros/roamer/commit/f2634a8))
- Internal refactoring and small fixes ([#a4dfb69](https://github.com/slipros/roamer/commit/a4dfb69))

### Fixed
- String formatter bug fixes ([#03c3b95](https://github.com/slipros/roamer/commit/03c3b95))

### Documentation
- Extended parser query documentation ([#08426d8](https://github.com/slipros/roamer/commit/08426d8))
- Updated README with latest features ([#8c20432](https://github.com/slipros/roamer/commit/8c20432))

## [1.15.0] - 2025-09-27

### Changed
- Switched from internal value conversion to external [assign](https://github.com/slipros/assign) library ([#74](https://github.com/slipros/roamer/pull/74))
  - Removed internal `value` package (float, integer, ptr, slice, string, time conversion functions)
  - Reduced codebase by ~9,000 lines of code
  - Improved maintainability by delegating type conversion to specialized library

### Added
- New configuration options leveraging assign library capabilities
- Extended cookie parser functionality
- Comprehensive extending documentation

## [1.14.1] - 2025-09-19

### Fixed
- Various bug fixes improving stability ([#73](https://github.com/slipros/roamer/pull/73))

### Changed
- Updated README with better examples
- Code formatting improvements with gofmt ([#72](https://github.com/slipros/roamer/pull/72))

## [1.14.0] - 2025-09-14

### Changed
- Major refactoring and improvements ([#71](https://github.com/slipros/roamer/pull/71))
- Updated CI dependencies ([#70](https://github.com/slipros/roamer/pull/70))
- Bumped `actions/configure-pages` from 4 to 5 ([#66](https://github.com/slipros/roamer/pull/66))

## [1.13.1] - 2025-09-07

### Fixed
- Fixed inconsistencies in codebase ([#65](https://github.com/slipros/roamer/pull/65))

### Changed
- General code improvements and refactoring ([#64](https://github.com/slipros/roamer/pull/64), [#63](https://github.com/slipros/roamer/pull/63))

## [1.13.0] - 2025-08-30

### Added
- GitHub Pages documentation site ([#62](https://github.com/slipros/roamer/pull/62))

### Changed
- Significant internal improvements and refactoring ([#60](https://github.com/slipros/roamer/pull/60))

### Dependencies
- Bumped `github.com/stretchr/testify` from 1.11.0 to 1.11.1 ([#61](https://github.com/slipros/roamer/pull/61))

## [1.12.1] - 2025-08-26

### Fixed
- Multipart decoder bug fixes ([#59](https://github.com/slipros/roamer/pull/59))

### Dependencies
- Bumped `github.com/stretchr/testify` from 1.10.0 to 1.11.0 ([#58](https://github.com/slipros/roamer/pull/58))
- Bumped `actions/checkout` from 4 to 5 ([#57](https://github.com/slipros/roamer/pull/57))
- Updated `github.com/go-chi/chi/v5` dependency ([#55](https://github.com/slipros/roamer/pull/55))

## [1.12.0] - 2025-05-25

### Changed
- **Performance**: Significant reduction in memory allocations ([#54](https://github.com/slipros/roamer/pull/54))
  - Optimized decoder implementations (JSON, XML, Form, Multipart)
  - Improved parser efficiency (Query, Header, Cookie, Path)
  - Added internal structure caching mechanism
  - Refactored formatters for better performance
  - Reduced overall allocations by optimizing string operations

### Security
- Fixed GitHub code scanning alert: Added permissions to workflows ([#53](https://github.com/slipros/roamer/pull/53))

## [1.11.3] - 2025-05-11

### Changed
- Code quality enhancements and refactoring ([#52](https://github.com/slipros/roamer/pull/52))

## [1.11.2] - 2025-02-10

### Fixed
- Fixed setting strings slice to slice of strings enumerations ([#47](https://github.com/slipros/roamer/pull/47))

## [1.11.1] - 2025-02-10

### Fixed
- Fixed appending elements to string enumeration slices ([#46](https://github.com/slipros/roamer/pull/46))

### Dependencies
- Bumped `coverallsapp/github-action` from 2.3.4 to 2.3.6 ([#45](https://github.com/slipros/roamer/pull/45))

## [1.11.0] - 2025-01-06

### Added
- Cookie parser implementation ([#43](https://github.com/slipros/roamer/pull/43))
  - New `parser.NewCookie()` for extracting data from HTTP cookies
  - Full support for cookie parsing with struct tags

### Documentation
- Updated README with cookie parser examples ([#44](https://github.com/slipros/roamer/pull/44))

## [1.10.0] - 2024-12-25

### Added
- Ability to obtain data from multiple specified headers ([#40](https://github.com/slipros/roamer/pull/40))
  - Header parser now supports fallback headers (try multiple header names)

### Dependencies
- Bumped `github.com/stretchr/testify` from 1.9.0 to 1.10.0 ([#39](https://github.com/slipros/roamer/pull/39))
- Bumped `coverallsapp/github-action` from 2.3.0 to 2.3.4 ([#38](https://github.com/slipros/roamer/pull/38))

## [1.9.0] - 2024-09-17

### Added
- Slice formatter/middleware for slice operations ([#35](https://github.com/slipros/roamer/pull/35))
  - Support for `unique`, `sort`, `limit`, `compact` operations on slices
  - Configurable via `slice` struct tag

## [1.8.0] - 2024-08-16

### Added
- String formatter for string manipulations ([#34](https://github.com/slipros/roamer/pull/34))
  - Operations: `trim_space`, `lower`, `upper`, `title`, `snake_case`, `camel_case`, etc.
  - Configurable via `string` struct tag

### Dependencies
- Bumped `golangci/golangci-lint-action` from 5 to 6 ([#33](https://github.com/slipros/roamer/pull/33))
- Bumped `coverallsapp/github-action` from 2.2.3 to 2.3.0 ([#32](https://github.com/slipros/roamer/pull/32))

## [1.7.2] - 2024-05-02

### Changed
- README improvements and documentation updates
- Minor bug fixes and dependency updates ([#31](https://github.com/slipros/roamer/pull/31))

### Dependencies
- Bumped `golangci/golangci-lint-action` from 4 to 5 ([#30](https://github.com/slipros/roamer/pull/30))

## [1.7.1] - 2024-03-10

### Changed
- Comprehensive test coverage improvements
- Minor bug fixes and stability enhancements ([#29](https://github.com/slipros/roamer/pull/29))

### Dependencies
- Updated various dependencies to latest versions

## [1.7.0] - 2024-03-08

### Added
- Experimental `FastStructField` parser for improved performance ([#28](https://github.com/slipros/roamer/pull/28))

### Changed
- **BREAKING**: Minimum Go version updated to 1.21
- Leverage Go 1.21 features for better performance

### Dependencies
- Bumped `github.com/stretchr/testify` from 1.8.4 to 1.9.0 ([#27](https://github.com/slipros/roamer/pull/27))
- Bumped `golangci/golangci-lint-action` from 3 to 4 ([#26](https://github.com/slipros/roamer/pull/26))

## [1.6.0] - 2024-02-11

### Changed
- Minor fixes and dependency updates ([#25](https://github.com/slipros/roamer/pull/25))

### Dependencies
- Bumped `actions/cache` from 3 to 4 ([#24](https://github.com/slipros/roamer/pull/24))
- Bumped `actions/setup-go` from 4 to 5 ([#23](https://github.com/slipros/roamer/pull/23))

## [1.5.0] - 2023-09-28

### Added
- Multipart/form-data decoder support ([#21](https://github.com/slipros/roamer/pull/21))
  - New `decoder.NewMultipartFormData()` for handling file uploads
  - Full support for multipart form parsing

### Changed
- **BREAKING**: AfterParse function signature changed ([#20](https://github.com/slipros/roamer/pull/20))
  - Provides more context for post-parse processing

## [1.4.0] - 2023-09-21

### Added
- Decode error type for better error handling ([#19](https://github.com/slipros/roamer/pull/19))
  - More specific error information during decoding process

### Dependencies
- Bumped `coverallsapp/github-action` from 2.2.1 to 2.2.3 ([#17](https://github.com/slipros/roamer/pull/17))
- Bumped `actions/checkout` from 3 to 4 ([#18](https://github.com/slipros/roamer/pull/18))

## [1.3.0] - 2023-09-09

### Added
- Support for `fmt.Stringer` interface ([#16](https://github.com/slipros/roamer/pull/16))
  - Custom string representation for parsed values

## [1.2.1] - 2023-08-16

### Fixed
- Fixed "assigned valued returns not supported" error ([#15](https://github.com/slipros/roamer/pull/15))

### Added
- Qodana static analysis integration ([#13](https://github.com/slipros/roamer/pull/13))

### Dependencies
- Bumped `coverallsapp/github-action` from 2.2.0 to 2.2.1 ([#14](https://github.com/slipros/roamer/pull/14))

## [1.2.0] - 2023-07-01

### Changed
- Enhanced form decoder tests with additional test cases
- Comprehensive refactoring of test suite ([#11](https://github.com/slipros/roamer/pull/11))

## [1.1.0] - 2023-06-24

### Added
- MIT License ([#10](https://github.com/slipros/roamer/pull/10))

## [1.0.0] - 2023-06-24

### Added
- Initial release with core functionality ([#1](https://github.com/slipros/roamer/pull/1))
- Query parameter parser (`parser.NewQuery()`)
- Header parser (`parser.NewHeader()`)
- Path parameter parser (`parser.NewPath()`)
- JSON decoder (`decoder.NewJSON()`)
- XML decoder (`decoder.NewXML()`)
- Form URL-encoded decoder (`decoder.NewFormURL()`)
- HTTP middleware support
- Struct tag-based configuration
- Type conversion and validation
- Router integrations:
  - Chi router (`pkg/chi`)
  - Gorilla Mux (`pkg/gorilla`)
- Comprehensive test suite
- Documentation and examples

---

## Router-Specific Releases

### pkg/httprouter

#### [v1.1.0] - 2025-01-06
- Added router lookup for httprouter path parser ([#42](https://github.com/slipros/roamer/pull/42))

#### [v1.0.0] - 2025-01-06
- Initial httprouter integration ([#41](https://github.com/slipros/roamer/pull/41))

### pkg/chi

#### [v1.2.0] - 2024-12-25
- Updated with multiple headers support

#### [v1.1.1] - 2024-03-10
- Tests and minor fixes

#### [v1.1.0] - 2024-02-11
- Minor improvements

#### [v1.0.0] - 2023-08-16
- Initial chi router integration

### pkg/gorilla

#### [v1.1.1] - 2024-03-10
- Tests and minor fixes

#### [v1.1.0] - 2024-02-11
- Minor improvements

#### [v1.0.0] - 2023-08-16
- Initial Gorilla Mux integration

[Unreleased]: https://github.com/slipros/roamer/compare/v1.16.0...HEAD
[1.16.0]: https://github.com/slipros/roamer/compare/v1.15.0...v1.16.0
[1.15.0]: https://github.com/slipros/roamer/compare/v1.14.1...v1.15.0
[1.14.1]: https://github.com/slipros/roamer/compare/v1.14.0...v1.14.1
[1.14.0]: https://github.com/slipros/roamer/compare/v1.13.1...v1.14.0
[1.13.1]: https://github.com/slipros/roamer/compare/v1.13.0...v1.13.1
[1.13.0]: https://github.com/slipros/roamer/compare/v1.12.1...v1.13.0
[1.12.1]: https://github.com/slipros/roamer/compare/v1.12.0...v1.12.1
[1.12.0]: https://github.com/slipros/roamer/compare/v1.11.3...v1.12.0
[1.11.3]: https://github.com/slipros/roamer/compare/v1.11.2...v1.11.3
[1.11.2]: https://github.com/slipros/roamer/compare/v1.11.1...v1.11.2
[1.11.1]: https://github.com/slipros/roamer/compare/v1.11.0...v1.11.1
[1.11.0]: https://github.com/slipros/roamer/compare/v1.10.0...v1.11.0
[1.10.0]: https://github.com/slipros/roamer/compare/v1.9.0...v1.10.0
[1.9.0]: https://github.com/slipros/roamer/compare/v1.8.0...v1.9.0
[1.8.0]: https://github.com/slipros/roamer/compare/v1.7.2...v1.8.0
[1.7.2]: https://github.com/slipros/roamer/compare/v1.7.1...v1.7.2
[1.7.1]: https://github.com/slipros/roamer/compare/v1.7.0...v1.7.1
[1.7.0]: https://github.com/slipros/roamer/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/slipros/roamer/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/slipros/roamer/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/slipros/roamer/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/slipros/roamer/compare/v1.2.1...v1.3.0
[1.2.1]: https://github.com/slipros/roamer/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/slipros/roamer/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/slipros/roamer/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/slipros/roamer/releases/tag/v1.0.0
