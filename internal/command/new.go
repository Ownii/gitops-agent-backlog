package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ownii/gitops-agent-backlog/internal/backlog"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

const specTemplate = `## Spec

<what & why>

## Acceptance Criteria

- [ ]
`

// New scaffolds a new ticket folder and returns its path. It does not commit.
func New(cwd, slug string) (string, error) {
	if !ticket.ValidSlug(slug) {
		return "", fmt.Errorf("invalid slug %q: must be lowercase kebab-case ([a-z0-9] words joined by -)", slug)
	}
	r, err := repo.Discover(cwd)
	if err != nil {
		return "", err
	}
	if err := EnsureGab(r); err != nil {
		return "", err
	}
	active, doneIDs, err := backlog.Load(r)
	if err != nil {
		return "", err
	}
	id := backlog.NextID(active, doneIDs)
	rank := backlog.NextRank(active)
	folder := ticket.FormatFolder(rank, id, slug)
	dir := filepath.Join(r.TicketsDir(), folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	m := &ticket.Meta{ID: id, Title: slug, Status: ticket.StatusTodo}
	if err := ticket.WriteMeta(filepath.Join(dir, "meta.yml"), m); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(specTemplate), 0o644); err != nil {
		return "", err
	}
	return dir, nil
}
