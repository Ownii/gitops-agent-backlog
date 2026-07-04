package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

func Start(cwd, id string) error {
	r, err := repo.Discover(cwd)
	if err != nil {
		return err
	}
	tdir, folder, err := TicketDirByID(r, id)
	if err != nil {
		return err
	}
	metaPath := filepath.Join(tdir, "meta.yml")
	m, err := ticket.ReadMeta(metaPath)
	if err != nil {
		return err
	}
	if m.Status != ticket.StatusPlanned {
		return fmt.Errorf("ticket %s is %q, must be %q to start", id, m.Status, ticket.StatusPlanned)
	}

	branch := "gab/" + folder.ID + "-" + folder.Slug
	wt := r.WorktreePath(folder.ID, folder.Slug)
	if err := os.MkdirAll(filepath.Dir(wt), 0o755); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "worktree", "add", "-b", branch, wt, "main"); err != nil {
		return err
	}

	// Materialize the statusless brief and commit it on the branch.
	brief, err := buildBrief(tdir, r.DoDPath())
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(wt, ".gab"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(wt, ".gab", "BRIEF.md"), brief, 0o644); err != nil {
		return err
	}
	if _, err := gitx.Run(wt, "add", ".gab/BRIEF.md"); err != nil {
		return err
	}
	if _, err := gitx.Run(wt, "commit", "-m", "gab: brief for "+id); err != nil {
		return err
	}

	// Set truth on main.
	m.Status = ticket.StatusInProgress
	m.Branch = branch
	if err := ticket.WriteMeta(metaPath, m); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "add", metaPath); err != nil {
		return err
	}
	_, err = gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("gab: %s in-progress", id))
	return err
}

// buildBrief concatenates spec.md, plan.md and the global DoD into one file.
func buildBrief(ticketDir, dodPath string) ([]byte, error) {
	var b []byte
	appendFile := func(path, heading string) error {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		b = append(b, []byte("<!-- "+heading+" -->\n")...)
		b = append(b, data...)
		b = append(b, '\n')
		return nil
	}
	if err := appendFile(filepath.Join(ticketDir, "spec.md"), "spec"); err != nil {
		return nil, err
	}
	if err := appendFile(filepath.Join(ticketDir, "plan.md"), "plan"); err != nil {
		return nil, err
	}
	if err := appendFile(dodPath, "definition-of-done"); err != nil {
		return nil, err
	}
	return b, nil
}
