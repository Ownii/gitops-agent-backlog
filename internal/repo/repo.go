package repo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
)

type Repo struct {
	Main string // worktree path that has `main` checked out
}

// Discover finds the main worktree starting from any working directory
// inside the repository (the main checkout or a feature worktree).
func Discover(cwd string) (*Repo, error) {
	main, err := mainWorktree(cwd)
	if err != nil {
		return nil, err
	}
	return &Repo{Main: main}, nil
}

// mainWorktree parses `git worktree list --porcelain` and returns the path
// of the worktree checked out on refs/heads/main.
func mainWorktree(cwd string) (string, error) {
	out, err := gitx.Run(cwd, "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	var curPath string
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			curPath = strings.TrimPrefix(line, "worktree ")
		case line == "branch refs/heads/main":
			return curPath, nil
		}
	}
	return "", fmt.Errorf("no worktree checked out on branch main found from %s", cwd)
}

func (r *Repo) GabDir() string     { return filepath.Join(r.Main, ".gab") }
func (r *Repo) TicketsDir() string { return filepath.Join(r.GabDir(), "tickets") }
func (r *Repo) DoneDir() string    { return filepath.Join(r.GabDir(), "done") }
func (r *Repo) DoDPath() string    { return filepath.Join(r.GabDir(), "definition-of-done.md") }

// WorktreePath is the deterministic location for a ticket's feature worktree.
func (r *Repo) WorktreePath(id, slug string) string {
	return filepath.Join(filepath.Dir(r.Main), ".gab-worktrees", id+"-"+slug)
}
