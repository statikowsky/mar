package cli

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed guide.md
var guideText string

func newGuideCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:     "guide",
		Aliases: []string{"g"},
		Short:   "Print the mar agent guide (workflow + command cheatsheet)",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON {
				return printJSON(cmd, map[string]any{"guide": guideText})
			}
			fmt.Fprint(cmd.OutOrStdout(), guideText)
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}
