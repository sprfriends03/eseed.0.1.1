# Task 1.1: Core Infrastructure Setup Implementation Plan

## Pre-Implementation Checklist

### Environment Prerequisites
- [x] Go 1.21 or later installed
    - [x] Verify with: `go version`
    - [x] Update if needed: `brew upgrade go` or download from golang.org
- [x] Development tools
    - [x] Install golangci-lint
    - [x] Configure IDE (VSCode/GoLand)

### Access Requirements
- [x] MongoDB Access
    - [x] Local MongoDB installation OR
    - [x] MongoDB Atlas account with connection string
    - [x] Required permissions for database creation
- [x] Redis Access
    - [x] Local Redis server OR
    - [x] Redis Cloud account with credentials
- [x] MinIO/S3 Access
    - [x] Local MinIO setup OR
    - [x] S3-compatible storage credentials

## Implementation Checklist

### 1.1.1 Go Project Structure Setup

#### 1.1.1.1 Project Structure Organization
- [x] Verify and organize directory structure following architecture:
    - [x] route/
        - [x] index.go (middleware and route registration)
        - [x] auth.go (authentication routes)
        - [x] member.go (member management)
        - [x] membership.go (membership management)
        - [x] plant.go (plant management)
        - [x] harvest.go (harvest management)
        - [x] storage.go (file storage)
        - [x] cms.go (admin routes)
    - [x] store/{db,rdb,storage}
    - [x] pkg/{oauth,ws,ecode,mail,enum}
    - [x] docs/
    - [x] scripts/
    - [x] tests/{unit,integration}

### 1.1.2 Build and Development Setup
- [x] Review and update Makefile targets if needed:
    - [x] build
    - [x] test
    - [x] lint
    - [x] run
- [x] Verify build system
    - [x] Run: `make build`
    - [x] Verify binary creation

#### 1.1.1.3 Code Quality Setup
- [x] Install and configure linters:
    - [x] golangci-lint
    - [x] Create .golangci.yml
- [x] Set up Git hooks:
    - [x] pre-commit for linting
    - [x] pre-push for tests
- [x] Configure IDE settings:
    - [x] Format on save
    - [x] Import organization
    - [x] Code style settings

#### 1.1.1.4 Error Handling Framework
- [x] Create error types in pkg/ecode
    - [x] Created domain-specific error codes in cannabis.go
    - [x] Added documentation for error codes
- [x] Implement error wrapping
    - [x] Added WithContext for adding context information
    - [x] Added WithStack for stack trace information
    - [x] Added WrapIf for conditional error wrapping
- [x] Set up error logging
    - [x] Implemented LogError with severity levels
    - [x] Added LogErrorWithContext for context-aware logging
    - [x] Created structured error formatting for logs
- [x] Create error response helpers
    - [x] Added WithDesc for convenient error description
    - [x] Created documentation and examples
    - [x] Added README.md with best practices

#### 1.1.1.5 Code Refactoring and Optimization
- [x] Refactor domain files to maximize code reuse
    - [x] Embedded repo struct in domain types
    - [x] Standardized MongoDB query patterns
    - [x] Unified error handling approach
- [x] Optimize database operations
    - [x] Refactored PlantType domain
    - [x] Refactored SeasonalCatalog domain
    - [x] Refactored Payment domain
    - [x] Refactored NFTRecord domain
- [x] Ensure consistent patterns across codebase

### 1.1.2 MongoDB Configuration

#### 1.1.2.1 Database Setup
- [x] Create store/db/index.go:
    - [x] Connection settings
    - [x] Timeout configurations
    - [x] Pool settings
- [x] Implement connection management
    - [x] Connection creation
    - [x] Health checks
    - [x] Graceful shutdown

#### 1.1.2.2 Schema Definition
- [x] Define collection schemas in store/db:
    - [x] Members
    - [x] Memberships
    - [x] PlantSlots
    - [x] Plants
    - [x] CareRecords
    - [x] Harvests
    - [x] SeasonalCatalog
    - [x] PlantType
    - [x] Payment
    - [x] NFTRecord
- [x] Create indexes for each collection
- [x] Document schema relationships

### 1.1.3 Redis Setup

#### 1.1.3.1 Redis Configuration
- [x] Configure connection pool in store/rdb/index.go:
    - [x] Set pool size
    - [x] Set timeout values
    - [x] Configure retry logic
- [x] Implement health checks
- [x] Set up error handling

#### 1.1.3.2 Cache Management
- [x] Define cache key patterns
- [x] Set up TTL policies
- [x] Configure invalidation rules
- [x] Implement cache helpers

#### 1.1.3.3 Session Handling
- [x] Configure session storage
- [x] Set up session middleware
- [x] Implement session cleanup

### 1.1.4 MinIO Configuration

#### 1.1.4.1 Storage Setup
- [x] Configure MinIO client in store/storage/index.go
- [x] Create required buckets:
    - [x] profile-images
    - [x] documents
    - [x] plant-images
    - [x] harvest-images
    - [x] kyc-documents
    - [x] nft-metadata
- [x] Set up bucket policies

#### 1.1.4.2 Access Control
- [x] Configure CORS
- [x] Set up bucket policies
- [x] Implement access controls
- [x] Configure encryption

#### 1.1.4.3 Management
- [x] Set up lifecycle rules
- [x] Configure versioning
- [x] Implement backup strategy

### Testing Framework

#### Setup Test Environment
- [x] Create test configuration
- [x] Set up test databases
- [x] Configure test helpers
- [x] Create mock services

#### Implementation Tests
- [x] Unit test templates
- [x] Integration test setup
- [x] API test framework
- [ ] Performance test suite

### Security Implementation

#### Basic Security Setup
- [ ] Configure TLS/SSL
- [ ] Set up CORS policies
- [x] Implement rate limiting
- [ ] Configure security headers

#### Access Control
- [x] Set up authentication framework using existing OAuth
- [x] Configure authorization using existing role system
- [x] Implement API security

### Documentation

#### Technical Documentation
- [ ] API documentation
- [ ] Database schemas
- [ ] Configuration guide
- [ ] Development setup guide

#### Operational Documentation
- [ ] Deployment procedures
- [ ] Monitoring setup
- [ ] Backup procedures
- [ ] Emergency procedures

## Validation Checklist

### Infrastructure Validation
- [x] Verify all services running
- [x] Check connectivity:
    - [x] MongoDB connection
    - [x] Redis connection
    - [x] MinIO access
- [x] Verify logging
- [ ] Test monitoring

### Security Validation
- [ ] Run security scan
- [x] Test rate limiting
- [ ] Verify TLS configuration
- [ ] Check CORS settings

### Performance Validation
- [ ] Run load tests
- [x] Check connection pools
- [x] Verify cache operation
- [x] Test file operations

## Next Steps
- [ ] Review implementation with team
- [ ] Schedule security audit
- [ ] Plan monitoring setup
- [ ] Prepare for authentication implementation
