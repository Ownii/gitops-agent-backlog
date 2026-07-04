package ticket

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	StatusTodo       = "todo"
	StatusPlanned    = "planned"
	StatusInProgress = "in-progress"
	StatusToVerify   = "to-verify"
)

type Meta struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	Status    string   `yaml:"status"`
	Priority  string   `yaml:"priority,omitempty"`
	DependsOn []string `yaml:"depends_on,omitempty"`
	Branch    string   `yaml:"branch,omitempty"`
}

func ReadMeta(path string) (*Meta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var m Meta
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &m, nil
}

func WriteMeta(path string, m *Meta) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

type Folder struct {
	Rank int
	ID   string
	Slug string
	Name string
}

var folderRe = regexp.MustCompile(`^(\d{3})-(T\d+)-([a-z0-9]+(?:-[a-z0-9]+)*)$`)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidSlug reports whether s is lowercase kebab-case ([a-z0-9] words joined by -).
func ValidSlug(s string) bool {
	return slugRe.MatchString(s)
}

func ParseFolder(name string) (Folder, error) {
	m := folderRe.FindStringSubmatch(name)
	if m == nil {
		return Folder{}, fmt.Errorf("invalid ticket folder name %q", name)
	}
	rank, _ := strconv.Atoi(m[1])
	return Folder{Rank: rank, ID: m[2], Slug: m[3], Name: name}, nil
}

func FormatFolder(rank int, id, slug string) string {
	return fmt.Sprintf("%03d-%s-%s", rank, id, slug)
}
