package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

func Complete(cwd, id string) error {
	r, err := repo.Discover(cwd)
	if err != nil {
		return err
	}
	tdir, _, err := TicketDirByID(r, id)
	if err != nil {
		return err
	}
	metaPath := filepath.Join(tdir, "meta.yml")
	m, err := ticket.ReadMeta(metaPath)
	if err != nil {
		return err
	}
	if m.Status != ticket.StatusInProgress {
		return fmt.Errorf("ticket %s is %q, must be %q to complete", id, m.Status, ticket.StatusInProgress)
	}
	if m.Branch == "" {
		return fmt.Errorf("ticket %s has no branch recorded", id)
	}

	// Guard: complete must run from inside the ticket's own worktree. Run from
	// main, it would find main clean, see no SUMMARY.md, and silently advance
	// the ticket to to-verify without any work having happened.
	cur, err := gitx.Run(cwd, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}
	if cur != m.Branch {
		return fmt.Errorf("complete must run from the ticket's worktree (branch %s), but the current directory is on %q; cd into the worktree and retry", m.Branch, cur)
	}

	// The feature worktree (cwd) must be clean.
	status, err := gitx.Run(cwd, "status", "--porcelain")
	if err != nil {
		return err
	}
	if status != "" {
		return fmt.Errorf("worktree has uncommitted changes; commit before completing:\n%s", status)
	}

	// commitPaths is the exact set of files the to-verify commit may touch —
	// never whatever the user may have staged on main in parallel.
	commitPaths := []string{metaPath}

	// Flow summary.md back to the truth on main (if the agent wrote one).
	src := filepath.Join(cwd, ".gab", "SUMMARY.md")
	if data, rerr := os.ReadFile(src); rerr == nil {
		summaryPath := filepath.Join(tdir, "summary.md")
		if err := os.WriteFile(summaryPath, data, 0o644); err != nil {
			return err
		}
		if _, err := gitx.Run(r.Main, "add", summaryPath); err != nil {
			return err
		}
		commitPaths = append(commitPaths, summaryPath)
	} else if !os.IsNotExist(rerr) {
		return rerr
	}

	// Set status to-verify on main and commit.
	m.Status = ticket.StatusToVerify
	if err := ticket.WriteMeta(metaPath, m); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "add", metaPath); err != nil {
		return err
	}
	// Pathspec commit: commit ONLY the gab-owned paths.
	commitArgs := append([]string{"commit", "-m", fmt.Sprintf("gab: %s to-verify", id), "--"}, commitPaths...)
	if _, err := gitx.Run(r.Main, commitArgs...); err != nil {
		return err
	}

	// Best-effort push of the feature branch.
	if gitx.HasRemote(cwd, "origin") {
		if _, err := gitx.Run(cwd, "push", "-u", "origin", m.Branch); err != nil {
			fmt.Fprintf(os.Stderr, "gab: warning: push of %s failed: %v\n  (status is saved on main; push the branch manually)\n", m.Branch, err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "gab: no origin remote; skipped push of %s\n", m.Branch)
	}
	return nil
}
