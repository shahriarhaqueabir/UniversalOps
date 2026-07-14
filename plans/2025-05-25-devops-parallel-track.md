# Sprint: DevOps Parallel Track - Phase 1

This sprint targets the core DevOps roadmap provided by the user. We will transition from basic "Read-Only" system visibility to actionable software operations, moving logic into dedicated domain files and expanding into CI/CD, IaC, and DORA metrics.

## Phase 1: Core Domain Refactoring
| ID | Ticket | Status | Priority | File Scope | DOD |
|----|--------|--------|----------|------------|-----|
| D-01 | **Extract Git Domain** | 🔲 TODO | High | `internal/devops/git.go` | - [ ] Move `GetGitSummary` and `findGitRepos` to `internal/devops/git.go` - [ ] Implement `GetGitLog` and `GetGitDiff` - [ ] Add unit tests |
| D-02 | **Extract Docker Domain** | 🔲 TODO | High | `internal/devops/docker.go` | - [ ] Move `GetContainers` and `GetDockerStatus` to `internal/devops/docker.go` - [ ] Implement `ControlContainer(id, action)` (Start/Stop/Restart) - [ ] Add unit tests |
| D-03 | **Extract K8s Domain** | 🔲 TODO | High | `internal/devops/kubernetes.go` | - [ ] Move `GetKubernetesStatus` to `internal/devops/kubernetes.go` - [ ] Implement `GetK8sResources(namespace, resourceType)` - [ ] Add unit tests |

## Phase 2: Actionable Operations
| ID | Ticket | Status | Priority | Area | DOD |
|----|--------|--------|----------|------|-----|
| D-04 | **Git Core Actions** | 🔲 TODO | High | Git | - [ ] Implement `Fetch`, `Pull`, and basic `Commit` - [ ] Verify safety gates (block dangerous git commands via shell module) |
| D-05 | **Build System Integration** | 🔲 TODO | Med | Build | - [ ] Implement detection for `pom.xml`, `package.json`, `go.mod` - [ ] Add one-click "Install Dependencies" and "Build" actions |
| D-06 | **IaC Detection (Terraform)** | 🔲 TODO | Med | IaC | - [ ] Detect `.tf` files and `.terraform` state - [ ] Implement `TerraformPlan` and `TerraformValidate` visualization |

## Phase 3: CI/CD & Observability
| ID | Ticket | Status | Priority | Focus | DOD |
|----|--------|--------|----------|-------|-----|
| D-07 | **GitHub Actions Integration** | 🔲 TODO | High | CI/CD | - [ ] Implement `GetWorkflowRuns` using `gh` CLI or API - [ ] Show pipeline history in UI |
| D-08 | **DORA Metrics Foundation** | 🔲 TODO | High | Metrics | - [ ] Implement SQLite schema for tracking deployments - [ ] Track `Deployment Frequency` based on successful git tags/builds |

## Phase 4: DevOps Diagnostics
| ID | Ticket | Status | Priority | Focus | DOD |
|----|--------|--------|----------|-------|-----|
| D-09 | **DevOps Health Check** | 🔲 TODO | Med | Diagnostics | - [ ] One-click check: Git Clean, Dependencies Installed, Build Green - [ ] Report "DevOps Score" |
