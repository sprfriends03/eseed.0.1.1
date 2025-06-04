# Task 1.1: Core Infrastructure Setup Implementation Plan

## Pre-Implementation Checklist

### Environment Prerequisites
- [ ] Go 1.21 or later installed
    - [ ] Verify with: `go version`
    - [ ] Update if needed: `brew upgrade go` or download from golang.org
- [ ] Development tools
    - [ ] Install golangci-lint
    - [ ] Configure IDE (VSCode/GoLand)

### Access Requirements
- [ ] MongoDB Access
    - [ ] Local MongoDB installation OR
    - [ ] MongoDB Atlas account with connection string
    - [ ] Required permissions for database creation
- [ ] Redis Access
    - [ ] Local Redis server OR
    - [ ] Redis Cloud account with credentials
- [ ] MinIO/S3 Access
    - [ ] Local MinIO setup OR
    - [ ] S3-compatible storage credentials

## Implementation Checklist

### 1.1.1 Go Project Structure Setup

#### 1.1.1.1 Project Structure Organization
- [ ] Verify and organize directory structure following architecture:
    - [ ] route/
        - [ ] index.go (middleware and route registration)
        - [ ] auth.go (authentication routes)
        - [ ] member.go (member management)
        - [ ] membership.go (membership management)
        - [ ] plant.go (plant management)
        - [ ] harvest.go (harvest management)
        - [ ] storage.go (file storage)
        - [ ] cms.go (admin routes)
    - [ ] store/{db,rdb,storage}
    - [ ] pkg/{oauth,ws,ecode,mail,enum}
    - [ ] docs/
    - [ ] scripts/
    - [ ] tests/{unit,integration}

### 1.1.2 Build and Development Setup
- [ ] Review and update Makefile targets if needed:
    - [ ] build
    - [ ] test
    - [ ] lint
    - [ ] run
- [ ] Verify build system
    - [ ] Run: `make build`
    - [ ] Verify binary creation

#### 1.1.1.3 Code Quality Setup
- [ ] Install and configure linters:
    - [ ] golangci-lint
    - [ ] Create .golangci.yml
- [ ] Set up Git hooks:
    - [ ] pre-commit for linting
    - [ ] pre-push for tests
- [ ] Configure IDE settings:
    - [ ] Format on save
    - [ ] Import organization
    - [ ] Code style settings

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
- [ ] Create store/db/index.go:
    - [ ] Connection settings
    - [ ] Timeout configurations
    - [ ] Pool settings
- [ ] Implement connection management
    - [ ] Connection creation
    - [ ] Health checks
    - [ ] Graceful shutdown

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
- [ ] Create test configuration
- [ ] Set up test databases
- [ ] Configure test helpers
- [ ] Create mock services

#### Implementation Tests
- [ ] Unit test templates
- [ ] Integration test setup
- [ ] API test framework
- [ ] Performance test suite

### Security Implementation

#### Basic Security Setup
- [ ] Configure TLS/SSL
- [ ] Set up CORS policies
- [ ] Implement rate limiting
- [ ] Configure security headers

#### Access Control
- [ ] Set up authentication framework using existing OAuth
- [ ] Configure authorization using existing role system
- [ ] Implement API security

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
- [ ] Verify all services running
- [ ] Check connectivity:
    - [ ] MongoDB connection
    - [ ] Redis connection
    - [ ] MinIO access
- [ ] Verify logging
- [ ] Test monitoring

### Security Validation
- [ ] Run security scan
- [ ] Test rate limiting
- [ ] Verify TLS configuration
- [ ] Check CORS settings

### Performance Validation
- [ ] Run load tests
- [ ] Check connection pools
- [ ] Verify cache operation
- [ ] Test file operations

## Next Steps
- [ ] Review implementation with team
- [ ] Schedule security audit
- [ ] Plan monitoring setup
- [ ] Prepare for authentication implementation
