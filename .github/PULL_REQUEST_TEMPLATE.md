---
name: Pull Request
about: Submit changes to Universal-Ops
title: "[PR] "
labels: []
assignees: ''
---

## Summary
<!-- Brief description of the changes (1–3 sentences) -->

## Related Issues
<!-- Closes #N, Fixes #N, etc. -->

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Refactor (no behavior change)
- [ ] Performance improvement
- [ ] CI/CD / tooling
- [ ] Documentation

## Testing
<!-- Describe the tests you ran or added -->

### Running Tests
- [ ] Backend tests pass (`go test ./internal/... -count=1`)
- [ ] Frontend tests pass (`npm test --prefix cmd/opsforall-gui/frontend`)
- [ ] TypeScript compiles clean (`tsc -b` — build mode; root tsconfig is project references, `--noEmit` is a false green)
- [ ] ESLint clean (`npm run lint --prefix cmd/opsforall-gui/frontend`)
- [ ] Race detection clean (`go test -race ./internal/... -count=1 -timeout 180s`)
- [ ] E2E smoke tests pass (`python tests/e2e/run_e2e_tests.py`, Windows only)
- [ ] Integration tests updated if storage/workflow/event code changed

### New Code Quality
- [ ] New backend code has ≥60% test coverage
- [ ] New frontend code has tests (component + store + hook as applicable)
- [ ] Edge cases and error paths are tested
- [ ] Accessibility checked if UI changed (axe scan recommended)
- [ ] Benchmarks added for performance-sensitive code

## Checklist
- [ ] Code follows project conventions (Go idioms, functional components)
- [ ] No cloud dependencies or telemetry added
- [ ] Tests added for new/changed functionality
- [ ] Documentation updated if applicable
- [ ] Coverage meets 60% threshold (project & patch)
