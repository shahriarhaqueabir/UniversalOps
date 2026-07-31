# Triage Labels

The skills speak in terms of five canonical triage roles. This file maps those roles to the actual label strings used in this repo's Jira issue tracker.

Jira labels are case-sensitive freeform strings. Apply them via the `labels` field when creating or updating issues.

| Role in mattpocock/skills | Label in Jira | Meaning |
|---------------------------|--------------|---------|
| `needs-triage` | `needs-triage` | Maintainer needs to evaluate this issue |
| `needs-info` | `needs-info` | Waiting on reporter for more information |
| `ready-for-agent` | `ready-for-agent` | Fully specified, ready for an AFK agent |
| `ready-for-human` | `ready-for-human` | Requires human implementation |
| `wontfix` | `wontfix` | Will not be actioned |

When a skill mentions a role (e.g. "apply the AFK-ready triage label"), use the corresponding label string from this table.

## Label operations in Jira

- **On create**: include `"labels": ["needs-triage"]` in the issue creation payload
- **Set labels (team-managed projects — KAN)**: pass a flat array — `"labels": ["ready-for-agent"]`. Team-managed (simplified) Jira projects do NOT support JSON `add`/`remove` syntax.
- **Company-managed projects**: use the standard JSON merge syntax — `"labels": [{"add": "ready-for-agent"}, {"remove": "needs-triage"}]`
- **Query by label**: `jql: "project = KAN AND labels = 'ready-for-agent'"`

## Wayfinder-specific labels

These are NOT triage roles — they are wayfinding type markers, but managed the same way:

| Label | Purpose |
|-------|---------|
| `wayfinder:map` | Marks the map issue for a wayfinding effort |
| `wayfinder:research` | Research (AFK) decision ticket |
| `wayfinder:prototype` | Prototype (HITL) decision ticket |
| `wayfinder:grilling` | Grilling (HITL) decision ticket |
| `wayfinder:task` | Task (HITL or AFK) decision ticket |

## Editing labels

Edit the label strings above to match whatever vocabulary your Jira instance actually uses. The canonical names should be kept in sync across `docs/agents/issue-tracker.md` and this file.
