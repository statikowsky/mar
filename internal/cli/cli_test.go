package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func runCmd(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func chdirTemp(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

func TestDocCreateAndShowJSON(t *testing.T) {
	chdirTemp(t)
	if _, err := runCmd(t, "", "init"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := runCmd(t, "# Body", "doc", "create", "--code", "auth",
		"--title", "Auth", "--type", "design", "--body", "-"); err != nil {
		t.Fatalf("doc create: %v", err)
	}
	out, err := runCmd(t, "", "doc", "show", "DOC-AUTH", "--json")
	if err != nil {
		t.Fatalf("doc show: %v", err)
	}
	var d struct {
		Code, Title, Body string
	}
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	// The store normalizes bodies to end with a newline.
	if d.Code != "DOC-AUTH" || d.Title != "Auth" || d.Body != "# Body\n" {
		t.Errorf("got %+v", d)
	}
}

func TestDocCreateDuplicateFails(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "doc", "create", "--code", "a", "--title", "A", "--type", "design")
	if _, err := runCmd(t, "", "doc", "create", "--code", "a", "--title", "B", "--type", "design"); err == nil {
		t.Fatal("expected duplicate to fail")
	}
}

func TestDocLinkShowsInJSON(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "doc", "create", "--code", "auth", "--title", "Auth", "--type", "design")
	runCmd(t, "", "task", "create", "--title", "Wire it", "--code", "1")
	if _, err := runCmd(t, "", "doc", "link", "DOC-AUTH", "T-1"); err != nil {
		t.Fatalf("doc link: %v", err)
	}
	out, err := runCmd(t, "", "task", "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "T-1") {
		t.Errorf("task missing from list: %s", out)
	}
	if _, err := runCmd(t, "", "task", "unlink", "T-1", "DOC-AUTH"); err != nil {
		t.Fatalf("task unlink: %v", err)
	}
}

// runMain drives the real Execute path (envelope handling included), capturing
// the command's own output, the JSON error envelope written to stderr, and the
// returned error.
func runMain(t *testing.T, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	root := newRootCmd()
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)
	err = executeWith(root, &errBuf)
	return out.String(), errBuf.String(), err
}

func decodeEnvelope(t *testing.T, stderr string) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal([]byte(stderr), &m); err != nil {
		t.Fatalf("stderr is not a JSON envelope: %v\n%q", err, stderr)
	}
	return m
}

func TestJSONErrorEnvelopeOnFailure(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	_, stderr, err := runMain(t, "", "doc", "show", "DOC-NOPE", "--json")
	if !errors.Is(err, ErrHandled) {
		t.Fatalf("err = %v, want ErrHandled", err)
	}
	m := decodeEnvelope(t, stderr)
	if m["error"] == "" {
		t.Errorf("envelope missing error message: %q", stderr)
	}
}

func TestJSONEnvelopeFlagEqualsForm(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	_, stderr, err := runMain(t, "", "doc", "show", "DOC-NOPE", "--json=true")
	if !errors.Is(err, ErrHandled) {
		t.Fatalf("--json=true: err = %v, want ErrHandled", err)
	}
	if m := decodeEnvelope(t, stderr); m["error"] == "" {
		t.Errorf("envelope missing error: %q", stderr)
	}
}

func TestNoJSONEnvelopeWithoutFlag(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	_, stderr, err := runMain(t, "", "doc", "show", "DOC-NOPE")
	if err == nil || errors.Is(err, ErrHandled) {
		t.Fatalf("err = %v, want the raw error (not ErrHandled)", err)
	}
	if stderr != "" {
		t.Errorf("no envelope expected without --json, got %q", stderr)
	}
}

// A flag *value* that happens to be the string "--json" must not trigger the
// envelope — the old os.Args scanner false-positived here.
func TestFlagValueNamedJSONDoesNotTrigger(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	// --title consumes "--json" as its value; the bool --json flag stays false.
	// Force a failure with a missing required flag (no --title... use a dup code).
	runCmd(t, "", "doc", "create", "--code", "dup", "--title", "A", "--type", "design")
	_, stderr, err := runMain(t, "", "doc", "create", "--code", "dup", "--title", "--json", "--type", "design")
	if err == nil {
		t.Fatal("expected duplicate-code failure")
	}
	if errors.Is(err, ErrHandled) || stderr != "" {
		t.Errorf("flag value '--json' must not trigger the envelope: err=%v stderr=%q", err, stderr)
	}
}

func TestCommandWantsJSONNilAndNoFlag(t *testing.T) {
	if commandWantsJSON(nil) {
		t.Error("nil command should not want JSON")
	}
	// serve has no --json flag.
	if commandWantsJSON(newServeCmd()) {
		t.Error("command without a --json flag should not want JSON")
	}
}

func TestNoEnvelopeOnSuccess(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	_, stderr, err := runMain(t, "", "doc", "create", "--code", "ok", "--title", "A", "--type", "design", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stderr != "" {
		t.Errorf("no envelope expected on success, got %q", stderr)
	}
}

func TestTaskCreateMoveAndBoardJSON(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "task", "create", "--title", "First", "--code", "1")
	runCmd(t, "", "task", "create", "--title", "Second", "--code", "2")
	if _, err := runCmd(t, "", "task", "move", "T-1", "--column", "Done"); err != nil {
		t.Fatalf("task move: %v", err)
	}
	out, err := runCmd(t, "", "board", "show", "--json")
	if err != nil {
		t.Fatalf("board show: %v", err)
	}
	var board []struct {
		Name  string `json:"name"`
		Tasks []struct {
			Code string `json:"code"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(out), &board); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	byName := map[string][]string{}
	for _, col := range board {
		for _, tk := range col.Tasks {
			byName[col.Name] = append(byName[col.Name], tk.Code)
		}
	}
	if len(byName["Done"]) != 1 || byName["Done"][0] != "T-1" {
		t.Errorf("Done = %v, want [T-1]", byName["Done"])
	}
	if len(byName["To do"]) != 1 || byName["To do"][0] != "T-2" {
		t.Errorf("To do = %v, want [T-2]", byName["To do"])
	}
}

func TestTaskCreateAndMovePlacementFlags(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "task", "create", "--title", "First", "--code", "1")
	runCmd(t, "", "task", "create", "--title", "Second", "--code", "2", "--first")
	runCmd(t, "", "task", "create", "--title", "Third", "--code", "3", "--after", "T-1")
	runCmd(t, "", "task", "create", "--title", "Fourth", "--code", "4", "--index", "2")
	if _, err := runCmd(t, "", "task", "move", "T-3", "--column", "To do", "--before", "T-2"); err != nil {
		t.Fatalf("task move --before: %v", err)
	}
	if _, err := runCmd(t, "", "task", "move", "T-1", "--column", "To do", "--last"); err != nil {
		t.Fatalf("task move --last: %v", err)
	}
	out, _ := runCmd(t, "", "task", "list", "--column", "To do", "--json")
	var tasks []struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal([]byte(out), &tasks); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	got := make([]string, len(tasks))
	for i, tk := range tasks {
		got[i] = tk.Code
	}
	want := []string{"T-3", "T-2", "T-4", "T-1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func TestTaskPlacementFlagsConflict(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "task", "create", "--title", "First", "--code", "1")
	if _, err := runCmd(t, "", "task", "create", "--title", "Bad", "--first", "--last"); err == nil {
		t.Fatal("expected create placement conflict")
	}
	if _, err := runCmd(t, "", "task", "move", "T-1", "--first", "--index", "1"); err == nil {
		t.Fatal("expected move placement conflict")
	}
}

func TestSingleLetterAliases(t *testing.T) {
	chdirTemp(t)
	if _, err := runCmd(t, "", "i"); err != nil {
		t.Fatalf("alias i (init): %v", err)
	}
	if _, err := runCmd(t, "", "t", "create", "--title", "Aliased"); err != nil {
		t.Fatalf("alias t (task): %v", err)
	}
	out, err := runCmd(t, "", "b", "show")
	if err != nil {
		t.Fatalf("alias b (board): %v", err)
	}
	if !strings.Contains(out, "Aliased") {
		t.Errorf("board via alias missing task: %s", out)
	}
	if _, err := runCmd(t, "", "d", "create", "--code", "x", "--title", "X", "--type", "design"); err != nil {
		t.Fatalf("alias d (doc): %v", err)
	}
	docs, err := runCmd(t, "", "d", "list")
	if err != nil {
		t.Fatalf("alias d list: %v", err)
	}
	if !strings.Contains(docs, "DOC-X") {
		t.Errorf("doc list via alias missing doc: %s", docs)
	}
}

func TestColumnAddBeforeAndMove(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	// add a column before the first one (the case the old --after couldn't do)
	if _, err := runCmd(t, "", "column", "add", "Under consideration", "--before", "To do"); err != nil {
		t.Fatalf("column add --before: %v", err)
	}
	out, _ := runCmd(t, "", "board", "show", "--json")
	var board []struct {
		Name string `json:"name"`
	}
	json.Unmarshal([]byte(out), &board)
	if len(board) == 0 || board[0].Name != "Under consideration" {
		t.Fatalf("first column = %+v, want Under consideration", board)
	}
	// move Done to the front
	if _, err := runCmd(t, "", "column", "move", "Done", "--before", "Under consideration"); err != nil {
		t.Fatalf("column move: %v", err)
	}
	out, _ = runCmd(t, "", "board", "show", "--json")
	board = nil
	json.Unmarshal([]byte(out), &board)
	if board[0].Name != "Done" {
		t.Errorf("after move, first column = %q, want Done", board[0].Name)
	}
}

func TestDocEditSetsDates(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "doc", "create", "--code", "auth", "--title", "Auth", "--type", "design")
	if _, err := runCmd(t, "", "doc", "edit", "DOC-AUTH", "--created", "2026-05-26", "--updated", "2026-06-03"); err != nil {
		t.Fatalf("doc edit dates: %v", err)
	}
	out, _ := runCmd(t, "", "doc", "show", "DOC-AUTH", "--json")
	var d struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	json.Unmarshal([]byte(out), &d)
	if !strings.HasPrefix(d.CreatedAt, "2026-05-26") {
		t.Errorf("CreatedAt = %q", d.CreatedAt)
	}
	if !strings.HasPrefix(d.UpdatedAt, "2026-06-03") {
		t.Errorf("UpdatedAt = %q", d.UpdatedAt)
	}
}

func TestTaskEditSetsDatesAndRejectsBad(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "", "task", "create", "--title", "X", "--code", "1")
	if _, err := runCmd(t, "", "task", "edit", "T-1", "--created", "2026-01-15"); err != nil {
		t.Fatalf("task edit date: %v", err)
	}
	out, _ := runCmd(t, "", "task", "show", "T-1", "--json")
	if !strings.Contains(out, "2026-01-15") {
		t.Errorf("created date not set: %s", out)
	}
	if _, err := runCmd(t, "", "task", "edit", "T-1", "--updated", "nope"); err == nil {
		t.Error("expected invalid-date error")
	}
}

func TestTaskCreateWithCode(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	if _, err := runCmd(t, "", "task", "create", "--title", "Manual", "--code", "39"); err != nil {
		t.Fatalf("task create --code: %v", err)
	}
	out, err := runCmd(t, "", "task", "show", "T-39", "--json")
	if err != nil {
		t.Fatalf("task show T-39: %v", err)
	}
	if !strings.Contains(out, `"code": "T-39"`) {
		t.Errorf("task not created with code T-39: %s", out)
	}
}

func TestDocImportConvertsHTML(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	htmlPath := filepath.Join(t.TempDir(), "report.html")
	htmlDoc := `<!DOCTYPE html><html><head><title>Imported Report</title></head>
<body><h2>Section</h2><p>Some <strong>bold</strong> text.</p>
<div class="callout warn"><p>Heads up.</p></div></body></html>`
	if err := os.WriteFile(htmlPath, []byte(htmlDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runCmd(t, "", "doc", "import", htmlPath, "--code", "rep", "--type", "report"); err != nil {
		t.Fatalf("doc import: %v", err)
	}
	out, err := runCmd(t, "", "doc", "show", "DOC-REP", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var d struct{ Code, Title, Body string }
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	if d.Title != "Imported Report" {
		t.Errorf("title = %q, want derived from <title>", d.Title)
	}
	if !strings.Contains(d.Body, "## Section") || !strings.Contains(d.Body, "**bold**") {
		t.Errorf("body not converted to markdown: %q", d.Body)
	}
	if !strings.Contains(d.Body, "> [!WARNING]") || !strings.Contains(d.Body, "> Heads up.") {
		t.Errorf("callout not mapped to alert: %q", d.Body)
	}
}

func TestDocImportExplicitTitleWins(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	htmlPath := filepath.Join(t.TempDir(), "r.html")
	os.WriteFile(htmlPath, []byte("<title>From HTML</title><body><p>x</p></body>"), 0o644)
	if _, err := runCmd(t, "", "doc", "import", htmlPath, "--code", "a", "--type", "design", "--title", "Override"); err != nil {
		t.Fatal(err)
	}
	out, _ := runCmd(t, "", "doc", "show", "DOC-A", "--json")
	if !strings.Contains(out, "Override") {
		t.Errorf("explicit --title should win: %s", out)
	}
}

func TestListenForServeFallsBackWhenPortBusy(t *testing.T) {
	// Occupy a port, then ask for it without --port explicit: should fall back.
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	busyPort := occupied.Addr().(*net.TCPAddr).Port

	ln, err := listenForServe("127.0.0.1", busyPort, false)
	if err != nil {
		t.Fatalf("expected fallback, got error: %v", err)
	}
	defer ln.Close()
	got := ln.Addr().(*net.TCPAddr).Port
	if got == busyPort {
		t.Errorf("expected a different port than the busy %d", busyPort)
	}
	if got == 0 {
		t.Errorf("fallback port should be a real assigned port")
	}
}

func TestListenForServeExplicitPortBusyErrors(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	busyPort := occupied.Addr().(*net.TCPAddr).Port

	if ln, err := listenForServe("127.0.0.1", busyPort, true); err == nil {
		ln.Close()
		t.Errorf("explicit busy port should error, not fall back")
	}
}

func TestListenForServeUsesRequestedPortWhenFree(t *testing.T) {
	probe, _ := net.Listen("tcp", "127.0.0.1:0")
	port := probe.Addr().(*net.TCPAddr).Port
	probe.Close() // free it

	ln, err := listenForServe("127.0.0.1", port, false)
	if err != nil {
		t.Fatalf("listen on free port: %v", err)
	}
	defer ln.Close()
	if got := ln.Addr().(*net.TCPAddr).Port; got != port {
		t.Errorf("got port %d, want requested %d", got, port)
	}
}
