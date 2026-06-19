package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/statikowsky/mar/internal/store"
	"github.com/statikowsky/mar/internal/version"
)

var osGetwd = os.Getwd

// ErrHandled signals that Execute already reported the failure (as a JSON
// error envelope on stderr); main should exit non-zero without printing it
// again.
var ErrHandled = errors.New("error already reported")

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "mar",
		Short:         "mar — a local Markdown documentation repository and kanban board",
		Version:       version.Display(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("mar {{.Version}}\n")
	root.AddCommand(newInitCmd())
	root.AddCommand(newDocCmd())
	root.AddCommand(newTaskCmd(), newBoardCmd(), newColumnCmd())
	root.AddCommand(newServeCmd(), newVersionCmd(), newGuideCmd())
	return root
}

func newVersionCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "version",
		Short: "Print the mar version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON {
				return printJSON(cmd, map[string]any{
					"version":  version.Version,
					"codename": version.Codename(version.Version),
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "mar %s\n", version.Display())
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func Execute() error {
	return executeWith(newRootCmd(), os.Stderr)
}

// executeWith runs root and, on failure, emits the documented
// {"error": "..."} envelope to stderr when the invoked command had --json set.
// "wants JSON" is read from the parsed flag (via the command ExecuteC returns),
// so it handles --json and --json=true and never mistakes a flag value of
// "--json" for the flag itself.
func executeWith(root *cobra.Command, stderr io.Writer) error {
	cmd, err := root.ExecuteC()
	if err == nil {
		return nil
	}
	if commandWantsJSON(cmd) {
		enc := json.NewEncoder(stderr)
		enc.Encode(map[string]string{"error": err.Error()})
		return ErrHandled
	}
	return err
}

// commandWantsJSON reports whether the executed command parsed --json as true.
// Commands without a --json flag (e.g. serve) report false.
func commandWantsJSON(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	f := cmd.Flags().Lookup("json")
	return f != nil && f.Value.String() == "true"
}

func openStore() (*store.Store, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	path, err := store.Discover(wd)
	if err != nil {
		return nil, err
	}
	return store.Open(path)
}

func readBody(cmd *cobra.Command, path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if path == "-" {
		b, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(b), nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read body file %s: %w", path, err)
	}
	return string(b), nil
}

func printJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
