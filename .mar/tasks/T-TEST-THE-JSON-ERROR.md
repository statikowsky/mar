---
title: Test the --json error envelope end-to-end
status: active
created: "2026-06-12T19:19:04.212715Z"
updated: "2026-06-12T22:37:52.177577Z"
---
From project review. `Execute()`/`wantsJSON()` (internal/cli/root.go:53-66) are 0% covered — the path that turns command failures into the documented {"error": ...} JSON envelope on stderr. `TestJSONErrorEnvelope` (cli_test.go:88) only tests the argv-parsing helper, not the envelope. Tests bypass Execute via newRootCmd().Execute(). Also: `wantsJSON` rescans os.Args for literal `--json` (misses --json=true, false-positives on flag values), and the empty-error sentinel protocol (root.go:17 + main.go:12) would be cleaner as a named sentinel with errors.Is.
