package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/statikowsky/mar/internal/store"
)

func newInitCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:     "init",
		Aliases: []string{"i"},
		Short:   "Create a mar store in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			s, err := store.Init(wd)
			if err != nil {
				return fmt.Errorf("init store: %w", err)
			}
			defer s.Close()
			if asJSON {
				return printJSON(cmd, map[string]any{"initialized": true, "path": ".mar/"})
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Initialized mar store in .mar/")
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}
