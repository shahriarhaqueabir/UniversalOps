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
- [ ] Backend tests pass (`go test ./internal/...`)
- [ ] Frontend tests pass (`npm test --prefix cmd/opsforall-gui/frontend`)
- [ ] TypeScript compiles clean (`tsc --noEmit`)
- [ ] ESLint clean (`npm run lint --prefix cmd/opsforall-gui/frontend`)
- [ ] E2E tests pass (`python tests/e2e/run_e2e_tests.py`, Windows only)

## Checklist
- [ ] Code follows project conventions (Go idioms, functional components)
- [ ] No cloud dependencies or telemetry added
- [ ] Tests added for new/changed functionality
- [ ] Documentation updated if applicable
