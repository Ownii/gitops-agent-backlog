package backlog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

type Ticket struct {
	Folder ticket.Folder
	Meta   *ticket.Meta
	Dir    string
}

// Load reads active tickets (sorted by rank) and the set of done ticket ids.
func Load(r *repo.Repo) ([]Ticket, map[string]bool, error) {
	active, err := loadDir(r.TicketsDir(), true)
	if err != nil {
		return nil, nil, err
	}
	sort.Slice(active, func(i, j int) bool { return active[i].Folder.Rank < active[j].Folder.Rank })

	doneTickets, err := loadDir(r.DoneDir(), false)
	if err != nil {
		return nil, nil, err
	}
	done := map[string]bool{}
	for _, d := range doneTickets {
		done[d.Folder.ID] = true
	}
	return active, done, nil
}

// loadDir lists ticket folders in dir. When readMeta is true, meta.yml is loaded.
func loadDir(dir string, readMeta bool) ([]Ticket, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Ticket
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		f, perr := ticket.ParseFolder(e.Name())
		if perr != nil {
			continue // ignore non-ticket folders
		}
		td := Ticket{Folder: f, Dir: filepath.Join(dir, e.Name())}
		if readMeta {
			m, merr := ticket.ReadMeta(filepath.Join(td.Dir, "meta.yml"))
			if merr != nil {
				return nil, merr
			}
			td.Meta = m
		}
		out = append(out, td)
	}
	return out, nil
}

func NextRank(active []Ticket) int {
	max := 0
	for _, t := range active {
		if t.Folder.Rank > max {
			max = t.Folder.Rank
		}
	}
	return max + 10
}

func NextID(active []Ticket, doneIDs map[string]bool) string {
	max := 0
	consider := func(id string) {
		if n, err := strconv.Atoi(strings.TrimPrefix(id, "T")); err == nil && n > max {
			max = n
		}
	}
	for _, t := range active {
		consider(t.Folder.ID)
	}
	for id := range doneIDs {
		consider(id)
	}
	return "T" + strconv.Itoa(max+1)
}

// Next returns the first ready planned ticket by rank, an explanation of any
// blocked planned tickets, and a non-nil error only on a dependency cycle.
func Next(active []Ticket, doneIDs map[string]bool) (*Ticket, []string, error) {
	if err := detectCycle(active); err != nil {
		return nil, nil, err
	}
	activeIDs := map[string]bool{}
	for i := range active {
		activeIDs[active[i].Meta.ID] = true
	}
	var blocked []string
	for i := range active {
		t := &active[i]
		if t.Meta.Status != ticket.StatusPlanned {
			continue
		}
		missing := unmetDeps(t.Meta.DependsOn, doneIDs)
		if len(missing) == 0 {
			return t, blocked, nil
		}
		// Flag dependency ids that exist neither among active tickets nor in
		// done/ — a typo or a deleted ticket would otherwise block the ticket
		// forever while looking like a normal in-progress wait.
		annotated := make([]string, len(missing))
		for j, d := range missing {
			if !activeIDs[d] && !doneIDs[d] {
				annotated[j] = d + " (unknown id — typo or deleted ticket?)"
			} else {
				annotated[j] = d
			}
		}
		blocked = append(blocked, fmt.Sprintf("%s blocked on %s", t.Meta.ID, strings.Join(annotated, ", ")))
	}
	return nil, blocked, nil
}

func unmetDeps(deps []string, doneIDs map[string]bool) []string {
	var missing []string
	for _, d := range deps {
		if !doneIDs[d] {
			missing = append(missing, d)
		}
	}
	return missing
}

// detectCycle reports a cycle in depends_on edges among active tickets.
// Dependencies pointing outside the active set (e.g. to done tickets) are ignored.
func detectCycle(active []Ticket) error {
	index := map[string]*Ticket{}
	for i := range active {
		index[active[i].Meta.ID] = &active[i]
	}
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[string]int{}
	var visit func(id string) error
	visit = func(id string) error {
		color[id] = gray
		for _, dep := range index[id].Meta.DependsOn {
			if _, ok := index[dep]; !ok {
				continue // dep not among active tickets
			}
			switch color[dep] {
			case gray:
				return fmt.Errorf("dependency cycle involving %s and %s", id, dep)
			case white:
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		color[id] = black
		return nil
	}
	for id := range index {
		if color[id] == white {
			if err := visit(id); err != nil {
				return err
			}
		}
	}
	return nil
}
