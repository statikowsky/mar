package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/statikowsky/mar/internal/version"
)

func decodeJSON(t *testing.T, out string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	return m
}

func TestTaskMoveJSON(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "task", "create", "--title", "First", "--code", "1")
	out, err := runCmd(t, "", "task", "move", "T-1", "--column", "Done", "--json")
	if err != nil {
		t.Fatalf("task move: %v", err)
	}
	m := decodeJSON(t, out)
	if m["code"] != "T-1" {
		t.Errorf("code = %v, want T-1\n%s", m["code"], out)
	}
}

func TestTaskRmJSON(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "task", "create", "--title", "First", "--code", "1")
	out, err := runCmd(t, "", "task", "rm", "T-1", "--force", "--json")
	if err != nil {
		t.Fatalf("task rm: %v", err)
	}
	m := decodeJSON(t, out)
	if m["deleted"] != true || m["code"] != "T-1" {
		t.Errorf("got %v\n%s", m, out)
	}
}

func TestTaskLinkUnlinkJSON(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "doc", "create", "--code", "auth", "--title", "Auth", "--type", "design")
	runCmd(t, "", "task", "create", "--title", "Wire it", "--code", "1")
	out, err := runCmd(t, "", "task", "link", "T-1", "DOC-AUTH", "--json")
	if err != nil {
		t.Fatalf("task link: %v", err)
	}
	m := decodeJSON(t, out)
	if m["linked"] != true || m["task"] != "T-1" || m["doc"] != "DOC-AUTH" {
		t.Errorf("link got %v\n%s", m, out)
	}
	out, err = runCmd(t, "", "task", "unlink", "T-1", "DOC-AUTH", "--json")
	if err != nil {
		t.Fatalf("task unlink: %v", err)
	}
	m = decodeJSON(t, out)
	if m["unlinked"] != true {
		t.Errorf("unlink got %v\n%s", m, out)
	}
}

func TestDocMoveArchiveRmJSON(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "doc", "create", "--code", "auth", "--title", "Auth", "--type", "design")
	out, err := runCmd(t, "", "doc", "move", "DOC-AUTH", "--code", "login", "--json")
	if err != nil {
		t.Fatalf("doc move: %v", err)
	}
	if m := decodeJSON(t, out); m["code"] != "DOC-LOGIN" {
		t.Errorf("move code = %v, want DOC-LOGIN\n%s", m["code"], out)
	}
	out, err = runCmd(t, "", "doc", "archive", "DOC-LOGIN", "--json")
	if err != nil {
		t.Fatalf("doc archive: %v", err)
	}
	if m := decodeJSON(t, out); m["archived"] != true || m["code"] != "DOC-LOGIN" {
		t.Errorf("archive got %v\n%s", m, out)
	}
	out, err = runCmd(t, "", "doc", "rm", "DOC-LOGIN", "--force", "--json")
	if err != nil {
		t.Fatalf("doc rm: %v", err)
	}
	if m := decodeJSON(t, out); m["deleted"] != true || m["code"] != "DOC-LOGIN" {
		t.Errorf("rm got %v\n%s", m, out)
	}
}

func TestDocLinkUnlinkJSON(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "doc", "create", "--code", "auth", "--title", "Auth", "--type", "design")
	runCmd(t, "", "task", "create", "--title", "Wire it", "--code", "1")
	out, err := runCmd(t, "", "doc", "link", "DOC-AUTH", "T-1", "--json")
	if err != nil {
		t.Fatalf("doc link: %v", err)
	}
	if m := decodeJSON(t, out); m["linked"] != true || m["doc"] != "DOC-AUTH" || m["task"] != "T-1" {
		t.Errorf("link got %v\n%s", m, out)
	}
	out, err = runCmd(t, "", "doc", "unlink", "DOC-AUTH", "T-1", "--json")
	if err != nil {
		t.Fatalf("doc unlink: %v", err)
	}
	if m := decodeJSON(t, out); m["unlinked"] != true {
		t.Errorf("unlink got %v\n%s", m, out)
	}
}

func TestColumnAddMoveRenameRmJSON(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	out, err := runCmd(t, "", "column", "add", "Review", "--after", "In progress", "--json")
	if err != nil {
		t.Fatalf("column add: %v", err)
	}
	if m := decodeJSON(t, out); m["name"] != "Review" {
		t.Errorf("add name = %v, want Review\n%s", m["name"], out)
	}
	out, err = runCmd(t, "", "column", "move", "Review", "--before", "To do", "--json")
	if err != nil {
		t.Fatalf("column move: %v", err)
	}
	if m := decodeJSON(t, out); m["moved"] != true || m["column"] != "Review" {
		t.Errorf("move got %v\n%s", m, out)
	}
	out, err = runCmd(t, "", "column", "rename", "Review", "Triage", "--json")
	if err != nil {
		t.Fatalf("column rename: %v", err)
	}
	if m := decodeJSON(t, out); m["renamed"] != true || m["from"] != "Review" || m["to"] != "Triage" {
		t.Errorf("rename got %v\n%s", m, out)
	}
	out, err = runCmd(t, "", "column", "rm", "Triage", "--json")
	if err != nil {
		t.Fatalf("column rm: %v", err)
	}
	if m := decodeJSON(t, out); m["removed"] != true || m["column"] != "Triage" {
		t.Errorf("rm got %v\n%s", m, out)
	}
}

func TestJSONKeysAreLowercase(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "doc", "create", "--code", "auth", "--title", "Auth", "--type", "design")
	runCmd(t, "notes", "task", "create", "--title", "Wire it", "--code", "1", "--body", "-")

	taskOut, _ := runCmd(t, "", "task", "show", "T-1", "--json")
	for _, want := range []string{"\"code\"", "\"title\"", "\"body\"", "\"column\"", "\"status\"", "\"created_at\"", "\"updated_at\""} {
		if !strings.Contains(taskOut, want) {
			t.Errorf("task JSON missing key %s:\n%s", want, taskOut)
		}
	}
	for _, bad := range []string{"\"Code\"", "\"Title\"", "\"CreatedAt\"", "\"ColumnID\"", "\"id\"", "\"column_id\"", "\"position\""} {
		if strings.Contains(taskOut, bad) {
			t.Errorf("task JSON has stale key %s:\n%s", bad, taskOut)
		}
	}

	docOut, _ := runCmd(t, "", "doc", "show", "DOC-AUTH", "--json")
	for _, want := range []string{"\"code\"", "\"title\"", "\"type\"", "\"body\"", "\"status\"", "\"created_at\"", "\"updated_at\""} {
		if !strings.Contains(docOut, want) {
			t.Errorf("doc JSON missing key %s:\n%s", want, docOut)
		}
	}

	boardOut, _ := runCmd(t, "", "board", "show", "--json")
	if strings.Contains(boardOut, "\"Code\"") || strings.Contains(boardOut, "\"Title\"") {
		t.Errorf("board JSON still has PascalCase task keys:\n%s", boardOut)
	}
}

func TestBoardJSONOmitsTaskBody(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "Lots of notes here", "task", "create", "--title", "First", "--code", "1", "--body", "-")

	boardOut, _ := runCmd(t, "", "board", "show", "--json")
	if strings.Contains(boardOut, "Lots of notes here") {
		t.Errorf("board JSON should not inline task body:\n%s", boardOut)
	}
	if strings.Contains(boardOut, "\"body\"") {
		t.Errorf("board JSON should omit the body key entirely:\n%s", boardOut)
	}

	// task show must still return the body.
	showOut, _ := runCmd(t, "", "task", "show", "T-1", "--json")
	if !strings.Contains(showOut, "Lots of notes here") {
		t.Errorf("task show should still include body:\n%s", showOut)
	}
}

func TestShowJSONIncludesLinks(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "doc", "create", "--code", "auth", "--title", "Auth", "--type", "design")
	runCmd(t, "", "task", "create", "--title", "Wire it", "--code", "1")
	runCmd(t, "", "task", "link", "T-1", "DOC-AUTH")

	taskOut, _ := runCmd(t, "", "task", "show", "T-1", "--json")
	var tk struct {
		Code string   `json:"code"`
		Docs []string `json:"docs"`
	}
	if err := json.Unmarshal([]byte(taskOut), &tk); err != nil {
		t.Fatalf("unmarshal task: %v\n%s", err, taskOut)
	}
	if len(tk.Docs) != 1 || tk.Docs[0] != "DOC-AUTH" {
		t.Errorf("task docs = %v, want [DOC-AUTH]\n%s", tk.Docs, taskOut)
	}

	docOut, _ := runCmd(t, "", "doc", "show", "DOC-AUTH", "--json")
	var d struct {
		Code  string   `json:"code"`
		Tasks []string `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(docOut), &d); err != nil {
		t.Fatalf("unmarshal doc: %v\n%s", err, docOut)
	}
	if len(d.Tasks) != 1 || d.Tasks[0] != "T-1" {
		t.Errorf("doc tasks = %v, want [T-1]\n%s", d.Tasks, docOut)
	}

	// An unlinked task omits the docs key entirely (no null).
	runCmd(t, "", "task", "create", "--title", "Solo", "--code", "2")
	soloOut, _ := runCmd(t, "", "task", "show", "T-2", "--json")
	if strings.Contains(soloOut, "docs") {
		t.Errorf("unlinked task should omit docs key:\n%s", soloOut)
	}
}

func TestTaskJSONIncludesColumnName(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "task", "create", "--title", "Wire it", "--code", "1")
	runCmd(t, "", "task", "move", "T-1", "--column", "In progress")

	showOut, _ := runCmd(t, "", "task", "show", "T-1", "--json")
	var tk struct {
		Column string `json:"column"`
	}
	if err := json.Unmarshal([]byte(showOut), &tk); err != nil {
		t.Fatalf("unmarshal show: %v\n%s", err, showOut)
	}
	if tk.Column != "In progress" {
		t.Errorf("show column = %q, want In progress\n%s", tk.Column, showOut)
	}

	listOut, _ := runCmd(t, "", "task", "list", "--json")
	var tasks []struct {
		Column string `json:"column"`
	}
	if err := json.Unmarshal([]byte(listOut), &tasks); err != nil {
		t.Fatalf("unmarshal list: %v\n%s", err, listOut)
	}
	if len(tasks) != 1 || tasks[0].Column != "In progress" {
		t.Errorf("list column = %v, want [In progress]\n%s", tasks, listOut)
	}
}

func TestDocUnarchiveJSON(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "doc", "create", "--code", "auth", "--title", "Auth", "--type", "design")
	runCmd(t, "", "doc", "archive", "DOC-AUTH")

	out, err := runCmd(t, "", "doc", "unarchive", "DOC-AUTH", "--json")
	if err != nil {
		t.Fatalf("doc unarchive: %v", err)
	}
	if m := decodeJSON(t, out); m["unarchived"] != true || m["code"] != "DOC-AUTH" {
		t.Errorf("got %v\n%s", m, out)
	}

	showOut, _ := runCmd(t, "", "doc", "show", "DOC-AUTH", "--json")
	if !strings.Contains(showOut, "\"status\": \"active\"") {
		t.Errorf("doc not active after unarchive:\n%s", showOut)
	}
}

func TestTaskArchiveUnarchiveJSON(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "task", "create", "--title", "Wire it", "--code", "1")

	out, err := runCmd(t, "", "task", "archive", "T-1", "--json")
	if err != nil {
		t.Fatalf("task archive: %v", err)
	}
	if m := decodeJSON(t, out); m["archived"] != true || m["code"] != "T-1" {
		t.Errorf("archive got %v\n%s", m, out)
	}
	if showOut, _ := runCmd(t, "", "task", "show", "T-1", "--json"); !strings.Contains(showOut, "\"status\": \"archived\"") {
		t.Errorf("task not archived:\n%s", showOut)
	}

	out, err = runCmd(t, "", "task", "unarchive", "T-1", "--json")
	if err != nil {
		t.Fatalf("task unarchive: %v", err)
	}
	if m := decodeJSON(t, out); m["unarchived"] != true || m["code"] != "T-1" {
		t.Errorf("unarchive got %v\n%s", m, out)
	}
	if showOut, _ := runCmd(t, "", "task", "show", "T-1", "--json"); !strings.Contains(showOut, "\"status\": \"active\"") {
		t.Errorf("task not active after unarchive:\n%s", showOut)
	}
}

func TestTaskListStatusFilter(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "task", "create", "--title", "Active one", "--code", "1")
	runCmd(t, "", "task", "create", "--title", "Old one", "--code", "2")
	runCmd(t, "", "task", "archive", "T-2")

	// Default lists active only.
	def, _ := runCmd(t, "", "task", "list", "--json")
	if !strings.Contains(def, "T-1") || strings.Contains(def, "T-2") {
		t.Errorf("default list should show only active T-1:\n%s", def)
	}
	arch, _ := runCmd(t, "", "task", "list", "--status", "archived", "--json")
	if !strings.Contains(arch, "T-2") || strings.Contains(arch, "T-1") {
		t.Errorf("archived list should show only T-2:\n%s", arch)
	}
}

func TestVersionCommand(t *testing.T) {
	out, err := runCmd(t, "", "version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(out, version.Version) {
		t.Errorf("version output %q missing %q", out, version.Version)
	}
	if !strings.Contains(out, version.Codename(version.Version)) {
		t.Errorf("version output %q missing codename %q", out, version.Codename(version.Version))
	}
}

func TestVersionCommandJSON(t *testing.T) {
	out, err := runCmd(t, "", "version", "--json")
	if err != nil {
		t.Fatalf("version --json: %v", err)
	}
	m := decodeJSON(t, out)
	if m["version"] != version.Version {
		t.Errorf("got %v, want version=%q", m, version.Version)
	}
	if m["codename"] != version.Codename(version.Version) {
		t.Errorf("got %v, want codename=%q", m, version.Codename(version.Version))
	}
}

func TestVersionFlag(t *testing.T) {
	out, err := runCmd(t, "", "--version")
	if err != nil {
		t.Fatalf("--version: %v", err)
	}
	if !strings.Contains(out, version.Version) {
		t.Errorf("--version output %q missing %q", out, version.Version)
	}
}

func TestInitJSON(t *testing.T) {
	chdirTemp(t)
	out, err := runCmd(t, "", "init", "--json")
	if err != nil {
		t.Fatalf("init --json: %v", err)
	}
	m := decodeJSON(t, out)
	if m["initialized"] != true {
		t.Errorf("got %v\n%s", m, out)
	}
}

func TestMutationJSONErrorEnvelope(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	// A failing mutating command with --json should still error (envelope is
	// emitted by Execute, not runCmd) and must not print a success object.
	out, err := runCmd(t, "", "task", "move", "T-999", "--column", "Done", "--json")
	if err == nil {
		t.Fatal("expected error for unknown task")
	}
	if strings.Contains(out, "\"code\"") {
		t.Errorf("should not emit a success object on error: %s", out)
	}
}
