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
- **Web app**: `backend/src/`, `backend/tests/`
- Paths adjusted based on plan.md structure

## Phase 3.1: Setup
- [x] T001 Create project structure per implementation plan (backend/src/models, services, handlers, templates, assets/css, tests/)
- [x] T002 Clone my-go-saas-template and initialize Buffalo project with dependencies
- [x] T003 [P] Configure linting and formatting tools (gofmt, go vet)

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**
- [x] T004 [P] Contract test for /auth/soundcloud in backend/tests/contract/test_auth_soundcloud.go
- [x] T005 [P] Contract test for /auth/callback in backend/tests/contract/test_auth_callback.go
- [x] T006 [P] Contract test for /feed in backend/tests/contract/test_feed.go
- [x] T007 [P] Contract test for /filter in backend/tests/contract/test_filter.go
- [x] T008 [P] Integration test for authentication flow in backend/tests/integration/test_auth_flow.go
- [x] T009 [P] Integration test for feed display in backend/tests/integration/test_feed_display.go
- [x] T010 [P] Integration test for filtering in backend/tests/integration/test_filtering.go
- [x] T011 [P] Integration test for error handling (API failure) in backend/tests/integration/test_error_handling.go
- [x] T012 [P] Integration test for performance (<2s) in backend/tests/integration/test_performance.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)
- [x] T013 [P] User model in backend/src/models/user.go
- [x] T014 [P] Track model in backend/src/models/track.go
- [x] T015 [P] Feed model in backend/src/models/feed.go
- [x] T016 SoundcloudService for API integration in backend/src/services/soundcloud_service.go
- [x] T017 FeedService for caching and filtering in backend/src/services/feed_service.go
- [x] T018 Auth handlers in backend/src/handlers/auth.go
- [x] T019 Feed handlers in backend/src/handlers/feed.go
- [x] T020 Login template in backend/src/templates/login.html
- [x] T021 Feed template in backend/src/templates/feed.html
- [x] T022 Filter template in backend/src/templates/filter.html
- [x] T023 Add Pico.css to backend/assets/css/

## Phase 3.4: Integration
- [x] T024 Connect services to PostgreSQL
- [x] T025 Add auth middleware
- [x] T026 Add logging

## Phase 3.5: Polish
- [x] T027 [P] Unit tests for models in backend/tests/unit/test_models.go
- [x] T028 [P] Unit tests for services in backend/tests/unit/test_services.go
- [x] T029 Performance optimization (<2s)
- [x] T030 Update docs
- [x] T031 Run quickstart tests

## Dependencies
- Setup (T001-T003) before everything
- Tests (T004-T012) before implementation (T013-T023)
- Models (T013-T015) before services (T016-T017)
- Services (T016-T017) before handlers (T018-T019)
- Handlers (T018-T019) before templates (T020-T022)
- T024 blocks T025, T026
- Implementation before polish (T027-T031)

## Parallel Example
```
# Launch T004-T007 together:
Task: "Contract test for /auth/soundcloud in backend/tests/contract/test_auth_soundcloud.go"
Task: "Contract test for /auth/callback in backend/tests/contract/test_auth_callback.go"
Task: "Contract test for /feed in backend/tests/contract/test_feed.go"
Task: "Contract test for /filter in backend/tests/contract/test_filter.go"

# Launch T013-T015 together:
Task: "User model in backend/src/models/user.go"
Task: "Track model in backend/src/models/track.go"
Task: "Feed model in backend/src/models/feed.go"

# Launch T027-T028 together:
Task: "Unit tests for models in backend/tests/unit/test_models.go"
Task: "Unit tests for services in backend/tests/unit/test_services.go"
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
    - Each endpoint → implementation task

2. **From Data Model**:
    - Each entity → model creation task [P]
    - Relationships → service layer tasks

3. **From User Stories**:
    - Each story → integration test [P]
    - Quickstart scenarios → validation tasks

4. **Ordering**:
    - Setup → Tests → Models → Services → Endpoints → Polish
    - Dependencies block parallel execution

## Validation Checklist
*GATE: Checked by main() before returning*

- [x] All contracts have corresponding tests
- [x] All entities have model tasks
- [x] All tests come before implementation
- [x] Parallel tasks truly independent
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task