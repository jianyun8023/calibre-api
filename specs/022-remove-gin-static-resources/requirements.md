# Requirements Document

## Introduction

With the frontend migration to Next.js (web-next) completed in spec 009-frontend-migration, the Gin backend no longer needs to serve static files. This specification defines the requirements for removing static resource mapping support from the Gin router, simplifying the backend architecture and eliminating unnecessary configuration.

## Glossary

- **Gin**: The Go web framework used for the backend API server
- **Static Resource Mapping**: Gin's functionality to serve static files (HTML, CSS, JS, images) from the filesystem
- **Next.js**: The React framework now handling all frontend serving independently
- **Backend**: The Go-based API server that provides REST endpoints
- **StaticDir**: Configuration parameter specifying the directory containing static files

## Requirements

### Requirement 1

**User Story:** As a backend developer, I want to remove static file serving from Gin, so that the backend focuses solely on API functionality and reduces configuration complexity.

#### Acceptance Criteria

1. WHEN the Gin router initializes THEN the system SHALL NOT configure any static file directory mappings
2. WHEN the Gin router initializes THEN the system SHALL NOT register routes for serving index.html or favicon.ico
3. WHEN the Gin router initializes THEN the system SHALL NOT configure a NoRoute handler for SPA fallback
4. WHEN an unmatched route is requested THEN the system SHALL return a 404 JSON response instead of serving HTML
5. WHEN the server starts THEN the system SHALL NOT require the StaticDir configuration parameter

### Requirement 2

**User Story:** As a system administrator, I want the configuration to reflect that static files are no longer served, so that deployment documentation is accurate and configuration is minimal.

#### Acceptance Criteria

1. WHEN the configuration is loaded THEN the system SHALL NOT validate the StaticDir parameter
2. WHEN the configuration is loaded THEN the system SHALL NOT set a default value for StaticDir
3. WHEN the configuration is documented THEN the system SHALL NOT reference static file serving capabilities
4. WHEN the server starts with missing StaticDir THEN the system SHALL start successfully without errors

### Requirement 3

**User Story:** As a developer, I want clean separation between frontend and backend, so that each service has a single, well-defined responsibility.

#### Acceptance Criteria

1. WHEN the backend code is reviewed THEN the system SHALL contain no references to serving HTML files
2. WHEN the backend code is reviewed THEN the system SHALL contain no references to serving static assets
3. WHEN the backend routes are listed THEN the system SHALL only show API endpoints and MCP endpoints
4. WHEN the setPages function is examined THEN the system SHALL either be removed or contain only API-related setup
