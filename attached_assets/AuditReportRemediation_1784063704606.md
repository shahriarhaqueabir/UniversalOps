#(AllOpsFull) — AuditReport Remediation Prompts

Source: 4-agent audit, 61 findings (12 CRITICAL / 20 HIGH / 14 MEDIUM / 15 LOW).
Assumption made: sequence is CRITICAL → HIGH → MEDIUM → LOW, one issue per commit, verify-before-fix on every item. Adjust if you want a different order (e.g. by module instead of severity).

---

## 0. Persistent rules — put in AGENTS.md / project system prompt

```
- Treat the uploaded audit as a set of CLAIMS, not facts. Every claim must be verified
  against the actual current source before any fix is written. Audits go stale and
  agents hallucinate line numbers — confirm file:line and reproduce the behavior first.
- One issue = one branch/commit. Never bundle unrelated fixes.
- No fix is "done" until: (a) the original defect is reproduced or confirmed statically,
  (b) the fix is applied, (c) a regression test or manual repro-check proves it's closed,
  (d) you've stated what you did NOT fix / any residual risk.
- Do not silently expand scope. If fixing X requires touching Y, stop and report before
  proceeding.
- Do not delete error handling to make code compile — surface it.
- If a finding cannot be reproduced or looks wrong, report "claim not verified" with
  evidence — do not fix a phantom issue or force-fit a fix to match the audit's wording.
- After each batch of fixes, run existing tests + linter + build for all 3 target OSes
  if CI supports it. Report failures verbatim, don't summarize them away.
```

---

## 1. Verification pass (run before touching any code)

```
Task: Verify the attached audit findings against the current Hawkward codebase.
Do NOT fix anything in this pass.

For each finding ID (SEC-1..7, IPC-1..3, FE-1, CROSS-1, H1..H20, M1..M14):
1. Open the cited file/line. Confirm it still exists and matches the described issue.
2. If line numbers have drifted, locate the actual current location.
3. Classify each as: CONFIRMED / PARTIALLY-CONFIRMED (issue exists but description is
   inaccurate — state how) / NOT-FOUND (code has changed or claim doesn't hold) /
   DUPLICATE (same root cause as another finding — name it).
4. For CONFIRMED items, note any additional call sites with the same defect that the
   audit missed (e.g. if SEC-1's unsanitized exec.Command pattern also appears
   elsewhere, list those files too).

Output a table: ID | status | file:line (current) | one-line evidence | related findings.
Stop here. Wait for my go-ahead before fixing anything.
```

---

## 2. Triage & sequencing (after verification table comes back)

```
Using the verification table from step 1:
- Group CONFIRMED findings by root cause, not by original ID (the audit itself flagged
  8 cross-cutting themes — command injection, data races, error swallowing, sandbox
  bypass, etc. Some "separate" findings are one fix).
- Propose a fix order that resolves shared root causes once (e.g. a single sanitization
  helper for SEC-1/2/3/4 rather than four separate patches) and flag any fix that
  touches a "god file" (H16 DevOps.go, frontend DevOps.tsx) since those carry higher
  regression risk and should be scoped tighter or split further.
- For each proposed batch, state: what breaks if this fix is wrong, and how you'll
  detect that before merging.
Wait for approval before starting batch 1.
```

---

## 3. Fix loop (repeat per approved batch)

```
Fix batch: [IDs].

1. Restate the confirmed defect in your own words (not the audit's wording) to prove
   you understand the mechanism, not just the label.
2. Show the minimal diff. Prefer the audit's suggested fix only if you've confirmed
   it's actually correct for this codebase — if not, propose and justify an alternative.
3. If the fix touches shared state (mutex additions, singleton access, etc.), check for
   other readers/writers of that state you may now be racing against or deadlocking.
4. Add/update a test that fails on the old code and passes on the new code, where the
   codebase's test setup allows it. If no test is feasible, say why and describe the
   manual verification you did instead.
5. Report: files changed, why, what you explicitly left alone and why, and any new
   risk introduced by the fix itself.
```

---

## 4. Self-audit checkpoint (run every ~5 fixes, and at the end)

```
Review your own last N commits against these audit categories, independent of the
original ticket list:
- Did any fix introduce a new unsynchronized global, a new raw exec.Command outside
  the sandbox wrapper, or a new swallowed error (result, _ := ...)?
- Did any "fix" just suppress a symptom (e.g. catch+ignore) rather than close the
  defect?
- Cross-check H16/H1-H20 "sandbox bypass" theme: did new code default to
  common.SandboxedCommand, or did it add another raw exec.Command path?
Report findings the same way you reported the original audit — don't grade your own
work generously.
```

---

Full findings list for reference/copy-paste is in the source doc — not reproduced here to keep this file short; point opencode at the original file directly for IDs and descriptions.