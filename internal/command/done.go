package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

func Done(cwd, id string) error {
	r, err := repo.Discover(cwd)
	if err != nil {
		return err
	}
	tdir, folder, err := TicketDirByID(r, id)
	if err != nil {
		return err
	}
	m, err := ticket.ReadMeta(filepath.Join(tdir, "meta.yml"))
	if err != nil {
		return err
	}
	if m.Status != ticket.StatusToVerify {
		return fmt.Errorf("ticket %s is %q, must be %q for done", id, m.Status, ticket.StatusToVerify)
	}
	if m.Branch == "" {
		return fmt.Errorf("ticket %s has no branch recorded", id)
	}

	// Squash-merge the branch into main (staged, not committed).
	if _, err := gitx.Run(r.Main, "merge", "--squash", m.Branch); err != nil {
		return err
	}
	// main owns .gab truth: drop any .gab changes the branch carried (e.g. BRIEF.md).
	if _, err := gitx.Run(r.Main, "reset", "-q", "--", ".gab"); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "checkout", "--", ".gab"); err != nil {
		return err
	}
	// Remove any brief file that was newly added by the branch (untracked after reset).
	_ = os.Remove(filepath.Join(r.GabDir(), "BRIEF.md"))
	if _, err := gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("feat: %s (%s)", m.Title, id)); err != nil {
		return err
	}

	// Archive the ticket folder to done/.
	dest := filepath.Join(r.DoneDir(), folder.Name)
	if _, err := gitx.Run(r.Main, "mv", tdir, dest); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("chore(gab): archive %s", id)); err != nil {
		return err
	}

	// Remove the worktree and delete the branch.
	wt := r.WorktreePath(folder.ID, folder.Slug)
	if _, err := gitx.Run(r.Main, "worktree", "remove", "--force", wt); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "branch", "-D", m.Branch); err != nil {
		return err
	}
	return nil
}
