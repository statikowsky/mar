package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

// anchors that prove the guide carries both workflow and command content.
// Matched case-insensitively so prose capitalization is not load-bearing.
var guideAnchors = []string{
	"check the board", // workflow marker
	"--json",          // the structured-output contract
	"task create",     // representative commands
	"doc create",
	"board show",
}

func guideContains(s, anchor string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(anchor))
}

func TestGuidePrintsMarkdown(t *testing.T) {
	out, err := runCmd(t, "", "guide")
	if err != nil {
		t.Fatalf("guide: %v", err)
	}
	for _, a := range guideAnchors {
		if !guideContains(out, a) {
			t.Errorf("guide output missing %q", a)
		}
	}
}

func TestGuideJSON(t *testing.T) {
	out, err := runCmd(t, "", "guide", "--json")
	if err != nil {
		t.Fatalf("guide --json: %v", err)
	}
	var payload struct {
		Guide string `json:"guide"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	if payload.Guide == "" {
		t.Fatal("guide field is empty")
	}
	for _, a := range guideAnchors {
		if !guideContains(payload.Guide, a) {
			t.Errorf("guide JSON missing %q", a)
		}
	}
}

func TestGuideAlias(t *testing.T) {
	out, err := runCmd(t, "", "g")
	if err != nil {
		t.Fatalf("g: %v", err)
	}
	if !guideContains(out, "check the board") {
		t.Errorf("alias output missing workflow content:\n%s", out)
	}
}

// TestGuideDocumentsEveryCommand fails if a top-level command is added without
// documenting it in the guide, keeping the cheatsheet from going stale.
func TestGuideDocumentsEveryCommand(t *testing.T) {
	out, _ := runCmd(t, "", "guide")
	root := newRootCmd()
	for _, c := range root.Commands() {
		name := c.Name()
		switch name {
		case "help", "completion": // cobra built-ins, not part of mar's surface
			continue
		}
		if !strings.Contains(out, name) {
			t.Errorf("guide does not mention command %q", name)
		}
	}
}
