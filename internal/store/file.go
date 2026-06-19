package store

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type taskMeta struct {
	Title   string `yaml:"title"`
	Status  string `yaml:"status"`
	Created string `yaml:"created"`
	Updated string `yaml:"updated"`
}

type docMeta struct {
	Title   string   `yaml:"title"`
	Type    string   `yaml:"type"`
	Status  string   `yaml:"status"`
	Created string   `yaml:"created"`
	Updated string   `yaml:"updated"`
	Tasks   []string `yaml:"tasks,omitempty"`
}

type boardFile struct {
	Columns []boardColumn `yaml:"columns"`
}

type boardColumn struct {
	Name  string   `yaml:"name"`
	Tasks []string `yaml:"tasks"`
}

func marshalEntity(meta any, body string) ([]byte, error) {
	fm, err := yaml.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}
	var b bytes.Buffer
	b.WriteString("---\n")
	b.Write(fm)
	b.WriteString("---\n")
	if body != "" {
		b.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			b.WriteByte('\n')
		}
	}
	return b.Bytes(), nil
}

// splitFrontmatter returns the YAML between the opening and closing ---
// lines and the verbatim body that follows.
func splitFrontmatter(raw []byte) ([]byte, string, error) {
	const sep = "---\n"
	s := string(raw)
	if !strings.HasPrefix(s, sep) {
		return nil, "", errors.New("missing frontmatter")
	}
	rest := s[len(sep):]
	end := strings.Index(rest, "\n"+sep)
	if end < 0 {
		return nil, "", errors.New("unterminated frontmatter")
	}
	return []byte(rest[:end+1]), rest[end+1+len(sep):], nil
}

func marshalTaskFile(m taskMeta, body string) ([]byte, error) { return marshalEntity(m, body) }

func parseTaskFile(raw []byte) (taskMeta, string, error) {
	fm, body, err := splitFrontmatter(raw)
	if err != nil {
		return taskMeta{}, "", err
	}
	var m taskMeta
	if err := yaml.Unmarshal(fm, &m); err != nil {
		return taskMeta{}, "", fmt.Errorf("parse frontmatter: %w", err)
	}
	if m.Status == "" {
		m.Status = "active"
	}
	return m, body, nil
}

func marshalDocFile(m docMeta, body string) ([]byte, error) { return marshalEntity(m, body) }

func parseDocFile(raw []byte) (docMeta, string, error) {
	fm, body, err := splitFrontmatter(raw)
	if err != nil {
		return docMeta{}, "", err
	}
	var m docMeta
	if err := yaml.Unmarshal(fm, &m); err != nil {
		return docMeta{}, "", fmt.Errorf("parse frontmatter: %w", err)
	}
	if m.Status == "" {
		m.Status = "active"
	}
	return m, body, nil
}

func marshalBoardFile(b boardFile) ([]byte, error) {
	for i := range b.Columns {
		if b.Columns[i].Tasks == nil {
			b.Columns[i].Tasks = []string{}
		}
	}
	out, err := yaml.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("marshal board: %w", err)
	}
	return out, nil
}

func parseBoardFile(raw []byte) (boardFile, error) {
	var b boardFile
	if err := yaml.Unmarshal(raw, &b); err != nil {
		return boardFile{}, fmt.Errorf("parse board: %w", err)
	}
	for i := range b.Columns {
		if b.Columns[i].Tasks == nil {
			b.Columns[i].Tasks = []string{}
		}
	}
	return b, nil
}
