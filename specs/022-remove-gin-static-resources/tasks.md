# Implementation Plan

- [ ] 1. Update main.go to remove static file serving
- [x] 1.1 Remove or simplify setPages() function
  - Remove `r.Static("/assets", conf.StaticDir+"/assets")` call
  - Remove route handlers for `/`, `/index`, `/favico.ico`
  - Replace NoRoute handler to return JSON 404 response with format: `{"error": "route not found", "path": "<path>"}`
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [x] 1.2 Write property test for NoRoute JSON response
  - **Property 1: NoRoute returns JSON 404**
  - **Validates: Requirements 1.4**
  - Generate random invalid route paths
  - Verify all responses are JSON with 404 status
  - Verify responses never contain HTML content

- [x] 1.3 Write unit test for router initialization
  - **Example 1: Router initialization excludes static routes**
  - **Validates: Requirements 1.1, 1.2, 1.3**
  - Initialize router and verify no static routes registered
  - Verify `/`, `/index`, `/favico.ico` routes are absent

- [ ] 2. Update configuration structure
- [x] 2.1 Remove StaticDir from Config struct
  - Edit `internal/calibre/types.go`
  - Remove `StaticDir string` field from Config struct
  - _Requirements: 2.1, 2.2_

- [x] 2.2 Remove StaticDir from configuration initialization
  - Edit `main.go` initConfig() function
  - Remove `viper.SetDefault("staticDir", "./static")` line
  - Remove StaticDir validation from validateConfig() function
  - _Requirements: 1.5, 2.4_

- [x] 2.3 Write unit test for configuration loading
  - **Example 2 & 3: Config loads without StaticDir**
  - **Validates: Requirements 1.5, 2.1, 2.2, 2.4**
  - Load config without StaticDir field
  - Verify no errors and no default value set
  - Verify server initialization succeeds

- [ ] 3. Verify route composition
- [x] 3.1 Write property test for route list composition
  - **Property 2: All registered routes are API or MCP routes**
  - **Validates: Requirements 3.3**
  - Get complete route list from initialized router
  - Verify each route starts with `/api/`, `/mcp/`, or equals `/ping`
  - Verify no routes match static file patterns

- [x] 4. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
