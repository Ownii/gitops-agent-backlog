package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

const defaultDoD = `# Definition of Done

Every ticket must satisfy these before moving to ` + "`to-verify`" + `:

- [ ] All automated tests pass.
- [ ] Linting/formatting is clean.
- [ ] No leftover TODOs related to this ticket.
- [ ] Public behaviour is documented where it changed.

Edit this file to match your project. The worktree agent must show evidence
(actual command output) that these are met before completing.
`

// EnsureGab creates the .gab skeleton and a default DoD if missing.
func EnsureGab(r *repo.Repo) error {
	for _, d := range []string{r.GabDir(), r.TicketsDir(), r.DoneDir()} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	if _, err := os.Stat(r.DoDPath()); os.IsNotExist(err) {
		if err := os.WriteFile(r.DoDPath(), []byte(defaultDoD), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// TicketDirByID returns the active ticket folder path and parsed folder for an id.
func TicketDirByID(r *repo.Repo, id string) (string, ticket.Folder, error) {
	entries, err := os.ReadDir(r.TicketsDir())
	if err != nil {
		return "", ticket.Folder{}, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		f, perr := ticket.ParseFolder(e.Name())
		if perr == nil && f.ID == id {
			return filepath.Join(r.TicketsDir(), e.Name()), f, nil
		}
	}
	return "", ticket.Folder{}, fmt.Errorf("no active ticket with id %s", id)
}
