package repo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
)

type Repo struct {
	Main  string // worktree path that has the trunk branch checked out
	Trunk string // the trunk branch name (main, master, ...)
}

// Discover finds the trunk worktree starting from any working directory
// inside the repository (the main checkout or a feature worktree).
func Discover(cwd string) (*Repo, error) {
	trunk, err := detectTrunk(cwd)
	if err != nil {
		return nil, err
	}
	main, err := worktreeOnBranch(cwd, trunk)
	if err != nil {
		return nil, err
	}
	return &Repo{Main: main, Trunk: trunk}, nil
}

// detectTrunk determines the repository's trunk branch. It prefers the remote's
// advertised default (origin/HEAD) and otherwise falls back to the first
// conventional trunk branch that exists locally. "Truth on the trunk" is the
// design principle — the branch *name* need not be "main".
func detectTrunk(cwd string) (string, error) {
	if out, err := gitx.Run(cwd, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		ref := strings.TrimSpace(out)
		if i := strings.LastIndex(ref, "/"); i >= 0 && i+1 < len(ref) {
			return ref[i+1:], nil
		}
	}
	for _, name := range []string{"main", "master", "trunk"} {
		if _, err := gitx.Run(cwd, "rev-parse", "--verify", "--quiet", "refs/heads/"+name); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("could not determine the trunk branch (looked for origin/HEAD, then main/master/trunk) from %s", cwd)
}

// worktreeOnBranch parses `git worktree list --porcelain` and returns the path
// of the worktree checked out on the given branch.
func worktreeOnBranch(cwd, branch string) (string, error) {
	out, err := gitx.Run(cwd, "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	target := "branch refs/heads/" + branch
	var curPath string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r") // tolerate CRLF line endings
		switch {
		case strings.HasPrefix(line, "worktree "):
			curPath = strings.TrimPrefix(line, "worktree ")
		case line == target:
			return curPath, nil
		}
	}
	return "", fmt.Errorf("trunk branch %q is not checked out in any worktree; check out %q in your main checkout so gab can find the truth (searched from %s)", branch, branch, cwd)
}

func (r *Repo) GabDir() string     { return filepath.Join(r.Main, ".gab") }
func (r *Repo) TicketsDir() string { return filepath.Join(r.GabDir(), "tickets") }
func (r *Repo) DoneDir() string    { return filepath.Join(r.GabDir(), "done") }
func (r *Repo) DoDPath() string    { return filepath.Join(r.GabDir(), "definition-of-done.md") }

// WorktreePath is the deterministic location for a ticket's feature worktree:
// a sibling of the repo, namespaced by repo name so two gab repos sharing a
// parent directory cannot collide on the same id+slug.
func (r *Repo) WorktreePath(id, slug string) string {
	return filepath.Join(filepath.Dir(r.Main), ".gab-worktrees", filepath.Base(r.Main), id+"-"+slug)
}
