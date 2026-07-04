package main

import (
	"fmt"
	"io"
	"os"
)

const usage = `usage: gab-helper <command> [args]

commands:
  new <slug>        scaffold a new ticket folder (status: todo)
  start <id>        create worktree + brief, set status in-progress
  complete <id>     flow summary back to main, set status to-verify, push
  done <id>         squash-merge, archive to done/, remove worktree
  next              print the id of the next ready ticket
`

func dispatch(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	// command cases are wired up in later tasks
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}

func main() {
	os.Exit(dispatch(os.Args[1:], os.Stdout, os.Stderr))
}
