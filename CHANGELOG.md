# Changelog

All notable changes to CHORM will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of CHORM library
- Struct-to-table mapping with reflection
- CRUD operations (Create, Read, Update, Delete)
- Query builder with fluent interface
- Aggregate functions support
- Window functions support
- Migration system
- Schema management utilities
- Cluster support for distributed tables
- Transaction support
- Connection pooling
- TLS and compression support
- Comprehensive test suite
- Docker and Docker Compose setup
- CI/CD pipeline with GitHub Actions
- Documentation and examples

### Features
- **Core ORM**: Complete ORM functionality for ClickHouse
- **Type Mapping**: Automatic Go to ClickHouse type conversion
- **Query Builder**: Fluent interface for building complex queries
- **Aggregates**: Support for ClickHouse analytical functions
- **Migrations**: Schema versioning and management
- **Clustering**: Distributed table operations
- **Performance**: Optimized batch inserts and connection pooling

### Technical Details
- Go 1.21+ compatibility
- ClickHouse driver integration
- Reflection-based struct mapping
- Parameterized query support
- Error handling with context
- Comprehensive logging and debugging

## [1.0.0] - 2024-01-01

### Added
- Initial release of CHORM library
- Complete ORM functionality for ClickHouse
- Struct-to-table mapping with reflection
- CRUD operations (Create, Read, Update, Delete)
- Query builder with fluent interface
- Aggregate functions support
- Window functions support
- Migration system
- Schema management utilities
- Cluster support for distributed tables
- Transaction support
- Connection pooling
- TLS and compression support
- Comprehensive test suite
- Docker and Docker Compose setup
- CI/CD pipeline with GitHub Actions
- Documentation and examples

### Features
- **Core ORM**: Complete ORM functionality for ClickHouse
- **Type Mapping**: Automatic Go to ClickHouse type conversion
- **Query Builder**: Fluent interface for building complex queries
- **Aggregates**: Support for ClickHouse analytical functions
- **Migrations**: Schema versioning and management
- **Clustering**: Distributed table operations
- **Performance**: Optimized batch inserts and connection pooling

### Technical Details
- Go 1.21+ compatibility
- ClickHouse driver integration
- Reflection-based struct mapping
- Parameterized query support
- Error handling with context
- Comprehensive logging and debugging

## [0.1.0] - 2024-01-01

### Added
- Basic struct mapping functionality
- Simple CRUD operations
- Connection management
- Basic query builder
- Initial test suite

### Changed
- Initial development version

### Deprecated
- None

### Removed
- None

### Fixed
- None

### Security
- None

---

## Version History

### Version 1.0.0
- **Release Date**: 2024-01-01
- **Status**: Stable
- **Go Version**: 1.21+
- **ClickHouse Version**: 22.0+

### Version 0.1.0
- **Release Date**: 2024-01-01
- **Status**: Development
- **Go Version**: 1.21+
- **ClickHouse Version**: 22.0+

## Migration Guide

### From 0.1.0 to 1.0.0

#### Breaking Changes
- None (first stable release)

#### New Features
- Complete ORM functionality
- Advanced query builder
- Aggregate and window functions
- Migration system
- Cluster support
- Transaction support

#### Deprecations
- None

#### Removals
- None

## Contributing

To contribute to this changelog:

1. Add your changes under the `[Unreleased]` section
2. Use the appropriate category:
   - `Added` for new features
   - `Changed` for changes in existing functionality
   - `Deprecated` for soon-to-be removed features
   - `Removed` for now removed features
   - `Fixed` for any bug fixes
   - `Security` for security-related changes

3. When releasing a new version:
   - Move `[Unreleased]` content to the new version
   - Update the version number and date
   - Add a new `[Unreleased]` section

## Links

- [GitHub Repository](https://github.com/forester/chorm)
- [Documentation](https://github.com/forester/chorm/docs)
- [Issues](https://github.com/forester/chorm/issues)
- [Releases](https://github.com/forester/chorm/releases) 