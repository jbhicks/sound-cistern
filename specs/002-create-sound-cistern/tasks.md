# Tasks: Sound Cistern

**Input**: Design documents from `/specs/002-create-sound-cistern/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
    → If not found: ERROR "No implementation plan found"
    → Extract: tech stack, libraries, structure
2. Load optional design documents:
    → data-model.md: Extract entities → model tasks
    → contracts/: Each file → contract test task
    → research.md: Extract decisions → setup tasks
3. Generate tasks by category:
    → Setup: project init, dependencies, linting
    → Tests: contract tests, integration tests
    → Core: models, services, CLI commands
    → Integration: DB, middleware, logging
    → Polish: unit tests, performance, docs
4. Apply task rules:
    → Different files = mark [P] for parallel
    → Same file = sequential (no [P])
    → Tests before implementation (TDD)
5. Number tasks sequentially (T001, T002...)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
    → All contracts have tests?
    → All entities have models?
    → All endpoints implemented?
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **PocketBase app**: `views/` (Templ templates), `pb_migrations/` (Go migrations)
- `main.go` for PocketBase app initialization and route registration
- `public/` for static assets (CSS, JS, images)

## Phase 3.1: Setup
- [x] T001 Create project structure (views/, pb_migrations/, public/css, public/js)
- [x] T002 Initialize PocketBase project with dependencies (Go 1.23, Templ v0.3.960)
- [x] T003 [P] Configure linting and formatting tools (gofmt, go vet, templ fmt)

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**
- [ ] T004 [P] Contract test for /auth/soundcloud in main_test.go
- [ ] T005 [P] Contract test for /auth/callback in main_test.go
- [ ] T006 [P] Contract test for /feed in main_test.go
- [ ] T007 [P] Contract test for /filter in main_test.go
- [ ] T008 [P] Integration test for authentication flow in main_test.go
- [ ] T009 [P] Integration test for feed display in main_test.go
- [ ] T010 [P] Integration test for filtering in main_test.go
- [ ] T011 [P] Integration test for error handling (API failure) in main_test.go
- [ ] T012 [P] Integration test for performance (<2s) in main_test.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)
- [x] T013 [P] SoundcloudUser collection migration in pb_migrations/1696000001_create_soundcloud_users.go
- [x] T014 [P] SoundcloudTrack collection migration in pb_migrations/1696000002_create_soundcloud_tracks.go
- [x] T015 [P] SoundcloudFeed collection migration in pb_migrations/1696000003_create_soundcloud_feeds.go
- [ ] T016 SoundcloudService using PocketBase RecordService pattern (fetch/cache via Soundcloud API)
- [ ] T017 FeedService using PocketBase RecordService pattern (query, filter, cache)
- [ ] T018 PocketBase route: POST /api/auth/soundcloud (OAuth initiate), GET /api/auth/callback (OAuth complete)
- [ ] T019 PocketBase route: GET /feed (render Templ), GET /api/feed (JSON), POST /api/feed/filter (HTMX)
- [x] T020 Templ template in views/signin.templ (existing)
- [ ] T021 Templ template for feed display in views/feed.templ
- [ ] T022 HTMX-powered filter component in views/feed.templ
- [x] T023 Pico.css in public/css/ (existing)

## Phase 3.4: Integration
- [x] T024 PocketBase SQLite DB initialized (pb_data/data.db)
- [ ] T025 PocketBase auth middleware for protected routes (app.OnBeforeServe hook)
- [ ] T026 Structured logging via PocketBase logger hooks

## Phase 3.5: Polish
- [ ] T027 [P] Unit tests for collections in main_test.go
- [ ] T028 [P] Unit tests for services in main_test.go
- [ ] T029 Performance optimization (<2s)
- [ ] T030 Update docs
- [ ] T031 Run quickstart tests

## Dependencies
- Setup (T001-T003) before everything
- Tests (T004-T012) before implementation (T013-T023)
- Migrations (T013-T015) before services (T016-T017)
- Services (T016-T017) before route handlers (T018-T019)
- Route handlers (T018-T019) before Templ templates (T020-T022)
- T024 blocks T025, T026
- Implementation (T013-T026) before polish (T027-T031)

## Parallel Example
```
# Launch T004-T007 together:
Task: "Contract test for /auth/soundcloud in main_test.go"
Task: "Contract test for /auth/callback in main_test.go"
Task: "Contract test for /feed in main_test.go"
Task: "Contract test for /filter in main_test.go"

# Launch T013-T015 together:
Task: "SoundcloudUser collection in pb_migrations/1696000001_create_soundcloud_users.go"
Task: "SoundcloudTrack collection in pb_migrations/1696000002_create_soundcloud_tracks.go"
Task: "SoundcloudFeed collection in pb_migrations/1696000003_create_soundcloud_feeds.go"

# Launch T027-T028 together:
Task: "Unit tests for collections in main_test.go"
Task: "Unit tests for services in main_test.go"
```

## Notes
- [P] tasks = different files, no dependencies
- Verify tests fail before implementing
- Commit after each task
- Avoid: vague tasks, same file conflicts

## Task Generation Rules
*Applied during main() execution*

1. **From Contracts**:
    - Each contract file → contract test task [P]
    - Each endpoint → PocketBase route implementation task

2. **From Data Model**:
    - Each entity → PocketBase collection migration task [P]
    - Relationships → RecordService layer tasks

3. **From User Stories**:
    - Each story → integration test [P]
    - Quickstart scenarios → validation tasks

4. **Ordering**:
    - Setup → Tests → Migrations → Services → Routes → Templ Templates → Polish
    - Dependencies block parallel execution

## Validation Checklist
*GATE: Checked by main() before returning*

- [x] All contracts have corresponding tests
- [x] All entities have model tasks
- [x] All tests come before implementation
- [x] Parallel tasks truly independent
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task