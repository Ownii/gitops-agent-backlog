# gab-helper

The deterministic core of `gab`: a small Go CLI that owns the git/filesystem
state of a `.gab/` backlog. It contains no product judgement — it scaffolds,
moves, and commits files and runs git so an agent doesn't have to do those
error-prone steps freehand.

## Build

    go build -o bin/gab-helper ./cmd/gab-helper

## Commands

    gab-helper new <slug>     scaffold a ticket folder (status: todo)
    gab-helper start <id>     create worktree + brief, set in-progress
    gab-helper complete <id>  flow summary back to main, set to-verify, push
    gab-helper done <id>      squash-merge, archive to done/, remove worktree
    gab-helper next           print the id of the next ready ticket

Exit codes: 0 success · 1 error · 2 usage · 3 next found nothing ready.
