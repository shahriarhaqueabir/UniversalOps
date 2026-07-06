# Onboarding Design

## Goal

A first-time user with zero terminal experience should be able to launch Hawkward and immediately understand:
1. What the application is
2. What it can do for them
3. How to navigate it
4. How to get help

No manual reading required.

## Flow

```
┌──────────────────────────────────────┐
│ Step 1: Welcome                      │
│ "👋 Welcome to Hawkward!"           │
│ Brief intro, no technical jargon     │
│ Bottom: [n]ext  [q]uit              │
├──────────────────────────────────────┤
│ Step 2: What is Hawkward?            │
│ "5 ops layers in one tool"          │
│ Icons + short description each      │
│ Bottom: [b]ack  [n]ext  [q]uit     │
├──────────────────────────────────────┤
│ Step 3: What You Can Do             │
│ Feature highlights with emojis      │
│ Focus on value, not implementation  │
│ Bottom: [b]ack  [n]ext  [q]uit     │
├──────────────────────────────────────┤
│ Step 4: Navigation at a Glance      │
│ Minimal key reference (just basics) │
│ ↑↓ enter, 1-5, ?, q                │
│ Bottom: [b]ack  [n]ext  [q]uit     │
├──────────────────────────────────────┤
│ Step 5: You're All Set!             │
│ "Press [Enter] to begin"           │
│ → Main Menu appears                 │
└──────────────────────────────────────┘
```

## Design Principles

1. **No assumptions** — The wizard assumes the user has never used a terminal application before
2. **Progress visibility** — Step indicator always visible (`Step 2/5`)
3. **Freedom to skip** — Any screen can be exited with `q` or `esc`
4. **Back navigation** — `[b]` goes to previous step if the user missed something
5. **Replayable** — The onboarding can be re-triggered from the help screen

## Persistence

After completing onboarding, a marker file is created:
- **Windows:** `%APPDATA%\hawkward\.onboarded`
- **Linux/macOS:** `~/.config/hawkward/.onboarded`

If this file exists, onboarding is skipped and the main menu opens directly.

## Future Enhancements

- Video/gif demo embedded in the onboarding
- Interactive tutorial mode (guided walkthrough of each ops layer)
- Language detection and localized onboarding
