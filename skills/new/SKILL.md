---
description: Create and rigorously refine a new gab ticket (spec + acceptance criteria) in the local .gab backlog. Use when the user wants to add a new task to work on with gab.
argument-hint: "[short title]"
allowed-tools: Bash(gab-helper *)
---

You are creating a new gab ticket. A ticket is a shared, **interpretation-free**
contract for one unit of work: anyone who reads it later — you, another agent, or
the user in three weeks — must know exactly what to build and how to tell it is
done, with nothing left to guess. Truth lives in `.gab/tickets/` on the `main`
branch; `gab-helper` handles the deterministic scaffolding.

Your goal is **complete clarity, not fast transcription.** Do NOT write the spec
from the user's first sentence — refine it with them until there is no room for
interpretation and every edge case is accounted for.

## 1. Ground yourself first
Before asking anything, look at the relevant part of the repo — the code this will
touch, existing tickets under `.gab/`, docs, and `.gab/definition-of-done.md` — so
your questions are specific and informed, not generic.

## 2. Refine with the user — one question at a time
Ask questions **one at a time** (never dump a list), and prefer concrete
multiple-choice options over open-ended prompts — they are easier to answer and
surface disagreement faster. Keep going until each of these is unambiguous:

- **Purpose / why:** what problem does this solve, for whom, and why now?
- **Scope — in and out:** state explicitly what is IN and what is OUT. Cut anything
  not needed for this unit of work (YAGNI) — a bloated ticket is a vague ticket.
- **Behaviour and edge cases:** walk the happy path AND the unhappy ones. Actively
  hunt edge cases and failure modes — empty/invalid input, boundaries and limits,
  missing dependencies, permissions, concurrency, error handling — name each one
  and decide the intended behaviour. Do not leave them implied.
- **Success criteria:** how will we *verify* it is done? Each must be specific and
  checkable, never a feeling.
- **Obvious dependencies:** if the work clearly needs another ticket done first, note
  it — but only in prose (see below). Leave the machine-readable `depends_on` to
  `/gab:plan`, which sets it once the technical order is clear.

When the user is vague ("handle errors gracefully", "make it fast", "and so on"),
push back: turn it into something concrete and testable before you accept it. Ask
"what should happen when …?" until the fuzzy edges are gone.

## 3. Confirm shared understanding (gate)
Before writing anything, play back a short summary — goal, scope (in/out), the
edge-case decisions, and the acceptance criteria — and ask the user to confirm or
correct it. Proceed only once they agree it is complete and correct.

## 4. Scaffold and write the ticket
1. Turn the confirmed title into a lowercase kebab-case slug — letters, digits and
   hyphens only (e.g. "OAuth Login" -> "oauth-login").
2. Run `gab-helper new <slug>`. It prints the created folder path and writes
   `meta.yml` (status: todo) and an empty `spec.md`. It does NOT commit.
3. Write the folder's `spec.md`:
   - `## Spec` — the what and the why, the scope boundaries (in / out), and any
     obvious dependency in prose (e.g. "depends on T4"). Do not set `depends_on`.
   - `## Acceptance Criteria` — a checklist of specific, verifiable outcomes that
     covers the happy path AND every edge case and failure mode you agreed on.
4. Set a human-readable `title:` in the ticket's `meta.yml` — the confirmed title as
   a real sentence (e.g. "OAuth login via Google"), distinct from the kebab-case slug
   in the folder name. (`gab-helper new` seeds `title` with the slug as a fallback;
   replace it with the readable one — later steps and the merge commit use it.)

## 5. Ambiguity self-review (before committing)
Re-read the spec as a stranger with no context and questionable judgement:
- Could any line be read two different ways? Pick one meaning and make it explicit.
- Any vague words ("appropriate", "properly", "fast", "etc.")? Replace them with
  concrete, checkable statements.
- Any edge case you discussed but did not write down? Add it.
- Anything in scope that is not actually needed? Cut it (YAGNI).
Fix issues inline until a first-time reader could implement it without asking you
a single clarifying question.

## 6. Commit and hand off
Commit on main. Stage the ticket folder AND the global Definition of Done (the
latter is only newly created on the very first ticket, but it must be versioned as
part of main's truth — do not leave it untracked):
`git add .gab/tickets/<folder> .gab/definition-of-done.md` then
`git commit -m "gab: new ticket <id> <slug>"`. Tell the user the ticket id and that
`/gab:plan <id>` is the next step.

Do not create a worktree or write an implementation plan here. This ticket captures
the WHAT, the WHY, and how we will know it is done — the HOW (implementation
approach) belongs to `/gab:plan`.
