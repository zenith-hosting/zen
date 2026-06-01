# Todo Example Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a new `examples/todo` app that demonstrates a tiny in-memory todo list with Go Fiber, Zen SSR, Preact, and Vite.

**Architecture:** Keep the example self-contained and parallel to `examples/basic`. The Go process owns task storage in memory, exposes simple form POST routes, and renders the SSR page with Zen. The Preact frontend only hydrates the page and renders the form/list UI; the browser uses normal submits and redirects for mutations.

**Tech Stack:** Go, Fiber v3, Zen SSR, Preact, Vite, TypeScript, Tailwind CSS

---

## Chunk 1: Scaffold the new example

**Files:**
- Create: `examples/todo/go.mod`
- Create: `examples/todo/package.json`
- Create: `examples/todo/main.go`
- Create: `examples/todo/frontend/package.json`
- Create: `examples/todo/frontend/tsconfig.json`
- Create: `examples/todo/frontend/vite.config.ts`
- Create: `examples/todo/frontend/index.html`
- Create: `examples/todo/frontend/src/app.css`
- Create: `examples/todo/frontend/src/entry-client.tsx`
- Create: `examples/todo/frontend/src/entry-server.tsx`
- Create: `examples/todo/frontend/src/pages.ts`
- Create: `examples/todo/frontend/src/pages/Todo.tsx`

- [ ] **Step 1: Write the new example files**

- [ ] **Step 2: Build the frontend**

Run: `pnpm --dir examples/todo/frontend build`
Expected: Vite completes client and SSR builds without errors.

- [ ] **Step 3: Build the Go app**

Run: `cd examples/todo && go build .`
Expected: The example builds a binary named `todo`.

- [ ] **Step 4: Commit**

```bash
git add examples/todo docs/superpowers/plans/2026-06-01-todo-example.md
git commit -m "feat: add todo example"
```
