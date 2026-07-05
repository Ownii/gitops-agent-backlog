package command

import (
	"fmt"
	"strings"

	"github.com/Ownii/gitops-agent-backlog/internal/backlog"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
)

// List renders the active backlog as one line per ticket, ordered by rank:
//
//	<rank>  <id>  <status>  <title>  (deps: ...)
//
// Done tickets live in .gab/done/ and are intentionally omitted — this is the
// working backlog. The output is empty when no active tickets exist.
func List(cwd string) (string, error) {
	r, err := repo.Discover(cwd)
	if err != nil {
		return "", err
	}
	active, _, err := backlog.Load(r)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, t := range active {
		fmt.Fprintf(&b, "%03d  %-4s  %-11s  %s", t.Folder.Rank, t.Meta.ID, t.Meta.Status, t.Meta.Title)
		if len(t.Meta.DependsOn) > 0 {
			fmt.Fprintf(&b, "  (deps: %s)", strings.Join(t.Meta.DependsOn, ", "))
		}
		b.WriteByte('\n')
	}
	return b.String(), nil
}
