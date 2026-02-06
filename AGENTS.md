## Project Memory and Context

You have access to long-term project memory through documentation files. This helps you maintain context across sessions and coordinate with other AI agents.

### On Starting a New Session

1. **Look for `docs/ai/index.md`** - This is the entry point for structured project documentation
   - If it exists, read it first for orientation
   - Then read `docs/ai/current-work.md` for active tasks and recent progress

3. **If it does not exist**, this may be a new project
   - Create `docs/ai/` structure for larger projects (see below)

### Updating Documentation

**For projects with `docs/ai/` structure:**

| File | Update When |
|------|-------------|
| `current-work.md` | Always - track active tasks, blockers, progress |
| `known-issues.md` | You discover bugs or workarounds |
| `decisions.md` | Making significant architectural choices |
| `patterns.md` | Establishing or discovering code patterns |
| `architecture.md` | System structure changes |

Keep `current-work.md` small and focused. Move completed items to permanent docs or delete them.

### Creating Documentation Structure

For new projects, create:

```
docs/ai/
├── index.md          # Entry point - what to read when
├── current-work.md   # Active tasks (ephemeral)
├── architecture.md   # System design (stable)
├── decisions.md      # Key decisions with rationale (stable)
├── patterns.md       # Code conventions (stable)
└── known-issues.md   # Bugs and workarounds (semi-stable)
```

Optionally add `AI_CONTEXT.md` files in key subdirectories for focused context.

### Guidelines

- These docs are for AI agents, not humans - optimize for quick context loading
- Be concise but provide enough context to resume work from scratch
- Remove stale information promptly
- Prefer structured documentation over lengthy prose
- Link to related docs rather than duplicating content
- When debugging, check `known-issues.md` first

### Using Git for Context (when allowed)

If git operations are permitted:
- Write meaningful commit messages explaining WHY, not just WHAT
- Use `git log --oneline -20` to see recent context
- Use `git diff` to see uncommitted changes
- Branch names should describe work in progress

---

### File Purposes

| File | Type | Content |
|------|------|---------|
| `index.md` | Stable | Entry point, project overview, what to read when |
| `current-work.md` | Ephemeral | Active tasks, blockers, recent changes - updated constantly |
| `architecture.md` | Stable | System design, component relationships - rarely changes |
| `decisions.md` | Append-only | ADR-lite records of key decisions and rationale |
| `patterns.md` | Stable | Code conventions, common patterns to follow |
| `known-issues.md` | Semi-stable | Bugs, workarounds, gotchas - updated as issues found/fixed |

## Example: Minimal index.md

```markdown
# AI Agent Context Index

Read this file first when starting a new session.

## Quick Start
1. Read `current-work.md` for active tasks
2. Run `npm run smoke` to verify app is healthy
3. Check `known-issues.md` if you hit unexpected behavior

## Project Overview
[One paragraph describing what this project is]

## Key Commands
[3-5 most common commands]

## Documentation
- `current-work.md` - Active tasks and progress
- `architecture.md` - System design
- `known-issues.md` - Bugs and workarounds
```

# Available Tools

If you find you don't have any tools that would be helpful to you, please let me know, and I will install them for you.

# Validation

You should build everything in such a way that you can test it directly, rather than relying on human eyes, ears, and hands. 

# Static Analysis

In a similar vein, use static analysis tools whenever appropriate to insure good coding practices and to insure this growing project remains manageable longterm.
