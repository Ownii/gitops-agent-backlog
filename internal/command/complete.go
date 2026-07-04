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

	// The feature worktree (cwd) must be clean.
	status, err := gitx.Run(cwd, "status", "--porcelain")
	if err != nil {
		return err
	}
	if status != "" {
		return fmt.Errorf("worktree has uncommitted changes; commit before completing:\n%s", status)
	}

	// Flow summary.md back to the truth on main (if the agent wrote one).
	src := filepath.Join(cwd, ".gab", "SUMMARY.md")
	if data, rerr := os.ReadFile(src); rerr == nil {
		if err := os.WriteFile(filepath.Join(tdir, "summary.md"), data, 0o644); err != nil {
			return err
		}
		if _, err := gitx.Run(r.Main, "add", filepath.Join(tdir, "summary.md")); err != nil {
			return err
		}
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
	if _, err := gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("gab: %s to-verify", id)); err != nil {
		return err
	}

	// Best-effort push of the feature branch.
	if gitx.HasRemote(cwd, "origin") {
		if _, err := gitx.Run(cwd, "push", "-u", "origin", m.Branch); err != nil {
			fmt.Printf("gab: warning: push of %s failed: %v\n  (status is saved on main; push the branch manually)\n", m.Branch, err)
		}
	} else {
		fmt.Printf("gab: no origin remote; skipped push of %s\n", m.Branch)
	}
	return nil
}
