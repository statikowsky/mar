package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/statikowsky/mar/internal/store"
)

func TestScratchpadCLIAddShowEditAndRemove(t *testing.T) {
	chdirTemp(t)
	if _, err := runCmd(t, "", "init"); err != nil {
		t.Fatal(err)
	}

	out, err := runCmd(t, "Capture this", "scratch", "add", "--text", "-", "--x", "25", "--y", "-10", "--color", "yellow", "--json")
	if err != nil {
		t.Fatalf("scratch add: %v", err)
	}
	var note store.ScratchNote
	if err := json.Unmarshal([]byte(out), &note); err != nil {
		t.Fatalf("unmarshal add: %v\n%s", err, out)
	}
	if note.ID != "S-1" || note.Text != "Capture this" || note.X != 25 || note.Y != -10 || note.Color != "yellow" {
		t.Fatalf("added note = %+v", note)
	}

	out, err = runCmd(t, "Reworked", "scratch", "edit", "S-1", "--text", "-", "--width", "340", "--x", "0", "--json")
	if err != nil {
		t.Fatalf("scratch edit: %v", err)
	}
	if err := json.Unmarshal([]byte(out), &note); err != nil {
		t.Fatalf("unmarshal edit: %v\n%s", err, out)
	}
	if note.Text != "Reworked" || note.Width != 340 || note.X != 0 || note.Y != -10 {
		t.Fatalf("edited note = %+v", note)
	}

	out, err = runCmd(t, "", "scratch", "show", "--json")
	if err != nil {
		t.Fatalf("scratch show: %v", err)
	}
	var pad store.Scratchpad
	if err := json.Unmarshal([]byte(out), &pad); err != nil {
		t.Fatalf("unmarshal show: %v\n%s", err, out)
	}
	if pad.Revision != 2 || len(pad.Notes) != 1 || pad.Notes[0].Text != "Reworked" {
		t.Fatalf("scratchpad = %+v", pad)
	}

	if _, err := runCmd(t, "", "scratch", "rm", "S-1"); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("scratch rm without force error = %v", err)
	}
	if _, err := runCmd(t, "", "scratch", "rm", "S-1", "--force"); err != nil {
		t.Fatalf("scratch rm: %v", err)
	}
}

func TestScratchpadCLIRejectsEmptyText(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	if _, err := runCmd(t, "", "scratch", "add", "--text", "-"); err == nil {
		t.Fatal("empty scratch note should fail")
	}
}

func TestScratchpadCLIPromotesNotes(t *testing.T) {
	chdirTemp(t)
	runCmd(t, "", "init")
	runCmd(t, "Build it\nTask details", "scratch", "add", "--text", "-")
	out, err := runCmd(t, "", "scratch", "promote", "S-1", "--task", "--json")
	if err != nil {
		t.Fatalf("promote task: %v", err)
	}
	if !strings.Contains(out, `"code": "T-BUILD-IT"`) {
		t.Fatalf("task promotion output = %s", out)
	}

	runCmd(t, "Reference title\nReference body", "scratch", "add", "--text", "-")
	out, err = runCmd(t, "", "scratch", "promote", "S-2", "--doc", "--code", "REF", "--type", "reference", "--json")
	if err != nil {
		t.Fatalf("promote doc: %v", err)
	}
	if !strings.Contains(out, `"code": "DOC-REF"`) {
		t.Fatalf("doc promotion output = %s", out)
	}
	if _, err := runCmd(t, "", "scratch", "promote", "S-2", "--task", "--doc"); err == nil {
		t.Fatal("promote with both kinds should fail")
	}
}
