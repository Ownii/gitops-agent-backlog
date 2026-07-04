package command

import (
	"github.com/Ownii/gitops-agent-backlog/internal/backlog"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
)

func Next(cwd string) (string, []string, error) {
	r, err := repo.Discover(cwd)
	if err != nil {
		return "", nil, err
	}
	active, doneIDs, err := backlog.Load(r)
	if err != nil {
		return "", nil, err
	}
	chosen, blocked, err := backlog.Next(active, doneIDs)
	if err != nil {
		return "", nil, err
	}
	if chosen == nil {
		return "", blocked, nil
	}
	return chosen.Meta.ID, blocked, nil
}
