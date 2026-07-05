package command

import (
	"fmt"

	"github.com/Ownii/gitops-agent-backlog/internal/backlog"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
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
	if chosen != nil {
		return chosen.Meta.ID, blocked, nil
	}
	// Nothing is ready. If no planned ticket was even blocked on dependencies,
	// explain why the backlog has nothing to start rather than staying silent.
	if len(blocked) == 0 {
		blocked = append(blocked, explainNothingReady(active))
	}
	return "", blocked, nil
}

// explainNothingReady describes why no ticket is startable when no planned
// ticket is dependency-blocked (e.g. everything is still in todo).
func explainNothingReady(active []backlog.Ticket) string {
	if len(active) == 0 {
		return "the backlog is empty — create a ticket with /gab:new"
	}
	todo := 0
	for _, t := range active {
		if t.Meta.Status == ticket.StatusTodo {
			todo++
		}
	}
	if todo > 0 {
		return fmt.Sprintf("%d ticket(s) still in 'todo' — run /gab:plan <id> to make them startable", todo)
	}
	return "no ticket is 'planned' and ready (in-progress and to-verify tickets are not startable)"
}
