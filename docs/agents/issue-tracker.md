# Issue tracker: Jira (Atlassian Cloud)

Issues and specs for this repo live as Jira issues on Atlassian Cloud. The AI interacts with Jira through the **Atlassian MCP server** (registered in `.vscode/mcp.json`), which provides native MCP tools for all Jira operations. No CLI is required.

## Jira project

- **Cloud ID**: `d9df71a2-df1a-4d8c-8b72-c93426971e56`
- **Project key**: `KAN` (universalops)
- **Base URL**: `https://hawkward.atlassian.net`
- **Issue types**: `Epic` (ID `10001`, wayfinder maps), `Story` (ID `10004`, feature work), `Task` (ID `10003`, chore/ops), `Bug` (ID `10006`, defects), `Feature` (ID `10005`), `Sub-task` (ID `10002`, child of a story/task)

### ⚠️ Team-managed project constraints (KAN)

The KAN project uses Jira's **team-managed (simplified)** project model, which differs from classic company-managed projects:

| Constraint | Team-managed (KAN) | Company-managed (classic) |
|-----------|-------------------|--------------------------|
| **Epic-child linking** | Use `parent` field: `{"parent": {"key": "KAN-X"}}` — `customfield_10014` does NOT exist in KAN | Use `customfield_10014` (Epic Link) |
| **Labels** | Flat string array only: `["label1", "label2"]` | Supports JSON add/remove: `[{"add": "label"}]` |
| **Available for Task type** | `parent` field (type: issuelink) is available | `customfield_10014` is the standard |

**Always check `getJiraIssueTypeMetaWithFields` before creating issues if unsure which fields are available.**

## Conventions

All Jira operations use the MCP tools provided by the Atlassian MCP server. The tool names follow the pattern `jira_*` (e.g. `jira_createIssue`, `jira_searchIssues`, `jira_addComment`).

### Workflow (KAN project)

The KAN project has a 4-status workflow with these transition IDs:

| Transition | Status | When to use |
|-----------|--------|-------------|
| Initial | To Do | Default on create |
| `21` | In Progress | Start working on an issue |
| `31` | In Review | Ready for review/feedback |
| `41` | Done | Work is complete |

All transitions are global and available from any status.

### Basic operations

| Operation | MCP Tool / Pattern |
|-----------|-------------------|
| **Create an issue** | `jira_createIssue` with `projectKey`, `summary`, `issueType`, `description` |
| **Read an issue** | `jira_getIssue` with `issueKey` |
| **Search issues** | `jira_searchIssues` with `jql` string |
| **Update an issue** | `jira_updateIssue` with `issueKey` and update fields |
| **Add a comment** | `jira_addComment` with `issueKey` and `body` |
| **Transition status** | `jira_transitionIssue` with `issueKey` and `transitionId` |
| **List projects** | `jira_getProjects` |
| **List issue types** | `jira_getIssueTypes` for a project |
| **List transitions** | `jira_getTransitions` for an issue (to find available workflow states) |
| **Get user info** | `jira_getCurrentUser` or `jira_searchUsers` |

### Label management

Jira labels are freeform strings. **How you apply them depends on project type:**

**Team-managed (KAN) — flat array only:**
- `"labels": ["needs-triage", "wayfinder:map"]` — works on create AND update
- JSON `{"add": "label"}` / `{"remove": "label"}` syntax does **NOT** work on KAN
- To replace labels on update, pass the full desired array

**Company-managed — standard JSON merge syntax:**
- `"labels": [{"add": "label-name"}, {"remove": "label-name"}]`

Use the canonical triage label strings from `docs/agents/triage-labels.md`.

### Example: Create an issue

```
jira_createIssue with:
  projectKey: "KAN"
  summary: "Add DNS-over-HTTPS fallback resolver"
  issueType: "Story"
  description: "## Background\n\nThe current DNS resolver uses raw UDP which is blocked on some networks..."
  labels: ["ready-for-agent", "netops"]
```

### Example: Search with JQL

```
jira_searchIssues with:
  jql: "project = KAN AND labels = 'ready-for-agent' AND assignee IS EMPTY ORDER BY created ASC"
```

## Wayfinding operations

Used by `/wayfinder`. The **map** is a single Jira issue with child issues as tickets, linked through Jira's native issue hierarchy.

### Map

A Jira issue of type **Epic** (or `Task` when Epic is unavailable), labelled `wayfinder:map`.

- **Create**: `jira_createIssue` with `issueType: "Epic"` (ID `10001`), `labels: ["wayfinder:map"]`, project = KAN
- **Body format**:

```
## Destination
<what reaching the end of this map looks like>

## Notes
<domain, skills, standing preferences>

## Decisions so far
<!-- one line per closed ticket: name + link + gist -->

## Not yet specified
<!-- fog — suspected questions not yet ticketable -->

## Out of scope
<!-- beyond the destination — won't graduate -->
```

### Child tickets

Child issues are linked to the map using Jira's **"Relates to"** issue link type and labelled `wayfinder:<type>` where type is one of:

- `wayfinder:research` — AFK (reading, investigation)
- `wayfinder:prototype` — HITL (build-to-decide)
- `wayfinder:grilling` — HITL (conversation to decide)
- `wayfinder:task` — HITL or AFK (do-to-decide)

**Body format**:

```
## Question
<the decision or investigation this ticket resolves>

## Blocked by
UOPS-123, UOPS-456 (or "None — can start immediately")
```

### Blocking

Use Jira's native **"Blocks" / "is blocked by"** issue link type. This renders visually in Jira's UI, making the frontier visible at a glance.

- **Add blocker**: `jira_linkIssues` with `outwardIssueKey` (the blocker), `inwardIssueKey` (the blocked), `linkType: "Blocks"`
- A ticket is **unblocked** when every issue linked via "is blocked by" is in a closed status (Done, Resolved, Closed)

### Frontier query

The frontier = open, unblocked, unassigned child issues of the map:

```
jql: "project = KAN AND issueType != Epic AND status not in (Done, Resolved, Closed) AND assignee IS EMPTY ORDER BY created ASC"
```

> **Team-managed note**: In simplified projects, parent-child relationships don't use the `"Epic Link"` custom field, so `parent IS NOT EMPTY` in JQL won't work as expected. Query by `issueType` or labels instead. For company-managed projects, use `'Epic Link' IS NOT EMPTY` to find child issues.

Then for each candidate, check blocking edges using `jira_getIssue` and inspect the `fields.issuelinks` array for any `"type": {"inward": "is blocked by"}` with an unresolved outward issue.

### Claim

Assign the issue to the driving developer:

```
jira_updateIssue with:
  issueKey: "UOPS-<n>"
  assignee: { "id": "<current-user-account-id>" }
```

Use `jira_getCurrentUser` to get the account ID first.

### Resolve

1. **Record answer**: `jira_addComment` with the resolution as the body
2. **Close**: `jira_transitionIssue` to a final status (Done / Resolved)
3. **Update map**: `jira_addComment` on the map issue with a pointer: `- [<ticket-name>](link) — <one-line gist>`

## When a skill says "publish to the issue tracker"

Create a Jira issue using the MCP `jira_createIssue` tool.

## When a skill says "fetch the relevant ticket"

Use `jira_getIssue` with the issue key (e.g. `UOPS-42`), including `?expand=renderedFields,comments` for full detail.

## Pull requests as a triage surface

**PRs as a request surface: no.** Use Jira exclusively for issue tracking. GitHub PRs are handled through GitHub's native review workflow and are not triaged as Jira issues.

## Credentials & authentication

Jira credentials are configured through the Atlassian MCP server (`.vscode/mcp.json`). The MCP server handles OAuth 2.1 or API token authentication transparently — no manual credential handling needed.
