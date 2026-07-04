package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Ownii/gitops-agent-backlog/internal/command"
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
	case "new":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: gab-helper new <slug>")
			return 2
		}
		dir, err := command.New(".", args[1])
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		fmt.Fprintln(stdout, dir)
		return 0
	case "start":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: gab-helper start <id>")
			return 2
		}
		if err := command.Start(".", args[1]); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}

func main() {
	os.Exit(dispatch(os.Args[1:], os.Stdout, os.Stderr))
}
