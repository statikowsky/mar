package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/statikowsky/mar/internal/store"
)

func newSearchCmd() *cobra.Command {
	var docs, tasks, asJSON bool
	var status, docType string
	c := &cobra.Command{
		Use:   "search TERM",
		Short: "Search docs and tasks by title and body",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			results, err := s.Search(args[0], store.SearchOpts{
				Docs: docs, Tasks: tasks, Status: status, Type: docType,
			})
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, results)
			}
			if len(results) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No matches.")
				return nil
			}
			for _, r := range results {
				ctx := r.Type // doc type, or task column
				if r.Kind == "task" {
					ctx = r.Column
				}
				text := r.Title
				if r.Field == "body" {
					text = r.Snippet
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-24s %s/%s  %s  %s\n", r.Code, r.Kind, ctx, r.Field, text)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&docs, "docs", false, "search documents only")
	c.Flags().BoolVar(&tasks, "tasks", false, "search tasks only")
	c.Flags().StringVar(&status, "status", "", "active|archived|all (default active)")
	c.Flags().StringVar(&docType, "type", "", "filter docs by type")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}
