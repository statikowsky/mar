package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/statikowsky/mar/internal/convert"
	"github.com/statikowsky/mar/internal/store"
)

func newDocCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "doc", Aliases: []string{"d"}, Short: "Manage documents"}
	cmd.AddCommand(
		newDocCreateCmd(), newDocImportCmd(), newDocShowCmd(), newDocListCmd(),
		newDocEditCmd(), newDocMoveCmd(), newDocArchiveCmd(), newDocUnarchiveCmd(),
		newDocRmCmd(), newDocLinkCmd(), newDocUnlinkCmd(),
	)
	return cmd
}

func newDocImportCmd() *cobra.Command {
	var code, title, docType string
	var asJSON bool
	c := &cobra.Command{
		Use:   "import FILE.html",
		Short: "Import an HTML file, converting it to Markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read %s: %w", args[0], err)
			}
			body, err := convert.HTMLToMarkdown(string(raw))
			if err != nil {
				return err
			}
			if title == "" {
				title = convert.DocumentTitle(string(raw))
			}
			if title == "" {
				return fmt.Errorf("no title found in %s; pass --title", args[0])
			}
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			d, err := s.CreateDoc(code, title, docType, body)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, d)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Imported %s\n", d.Code)
			return nil
		},
	}
	c.Flags().StringVar(&code, "code", "", "short code, e.g. AUTH (required)")
	c.Flags().StringVar(&title, "title", "", "document title (default: from HTML <title> or <h1>)")
	c.Flags().StringVar(&docType, "type", "", "design|analysis|plan|report|board|reference|tooling (required)")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	c.MarkFlagRequired("code")
	c.MarkFlagRequired("type")
	return c
}

func newDocCreateCmd() *cobra.Command {
	var code, title, docType, bodyPath string
	var asJSON bool
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a document",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			body, err := readBody(cmd, bodyPath)
			if err != nil {
				return err
			}
			d, err := s.CreateDoc(code, title, docType, body)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, d)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", d.Code)
			return nil
		},
	}
	c.Flags().StringVar(&code, "code", "", "short code, e.g. AUTH (required)")
	c.Flags().StringVar(&title, "title", "", "document title (required)")
	c.Flags().StringVar(&docType, "type", "", "design|analysis|plan|report|board|reference|tooling (required)")
	c.Flags().StringVar(&bodyPath, "body", "", "markdown body file, or - for stdin")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	c.MarkFlagRequired("code")
	c.MarkFlagRequired("title")
	c.MarkFlagRequired("type")
	return c
}

func newDocShowCmd() *cobra.Command {
	var asJSON bool
	var render string
	c := &cobra.Command{
		Use:   "show DOC-CODE",
		Short: "Show a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			d, err := s.GetDoc(args[0])
			if err != nil {
				return err
			}
			backlinks, err := s.Backlinks(d.Code)
			if err != nil {
				return err
			}
			if asJSON {
				tasks, err := s.TaskCodesForDoc(d.Code)
				if err != nil {
					return err
				}
				return printJSON(cmd, struct {
					store.Doc
					Tasks     []string         `json:"tasks,omitempty"`
					Backlinks []store.Backlink `json:"backlinks,omitempty"`
				}{d, tasks, backlinks})
			}
			if render == "md" {
				fmt.Fprintln(cmd.OutOrStdout(), d.Body)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  [%s/%s]\nUpdated %s\n\n%s\n",
				d.Code, d.Title, d.Type, d.Status, d.UpdatedAt, d.Body)
			if len(backlinks) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "\nReferenced by:")
				for _, b := range backlinks {
					fmt.Fprintf(cmd.OutOrStdout(), "  %-14s %s\n", b.Code, b.Title)
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	c.Flags().StringVar(&render, "render", "", "md to print raw markdown body")
	return c
}

func newDocListCmd() *cobra.Command {
	var docType, status string
	var asJSON bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List documents",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			docs, err := s.ListDocs(docType, status)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, docs)
			}
			for _, d := range docs {
				fmt.Fprintf(cmd.OutOrStdout(), "%-14s %-10s %s\n", d.Code, d.Type, d.Title)
			}
			return nil
		},
	}
	c.Flags().StringVar(&docType, "type", "", "filter by type")
	c.Flags().StringVar(&status, "status", "", "active|archived")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newDocEditCmd() *cobra.Command {
	var title, docType, bodyPath, created, updated string
	var asJSON bool
	c := &cobra.Command{
		Use:   "edit DOC-CODE",
		Short: "Edit a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			var tp, ty, bp *string
			if cmd.Flags().Changed("title") {
				tp = &title
			}
			if cmd.Flags().Changed("type") {
				ty = &docType
			}
			if cmd.Flags().Changed("body") {
				body, err := readBody(cmd, bodyPath)
				if err != nil {
					return err
				}
				bp = &body
			}
			d, err := s.EditDoc(args[0], tp, ty, bp)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("created") || cmd.Flags().Changed("updated") {
				var cp, up *string
				if cmd.Flags().Changed("created") {
					cp = &created
				}
				if cmd.Flags().Changed("updated") {
					up = &updated
				}
				d, err = s.SetDocDates(d.Code, cp, up)
				if err != nil {
					return err
				}
			}
			if asJSON {
				return printJSON(cmd, d)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", d.Code)
			return nil
		},
	}
	c.Flags().StringVar(&title, "title", "", "new title")
	c.Flags().StringVar(&docType, "type", "", "new type")
	c.Flags().StringVar(&bodyPath, "body", "", "new body file, or -")
	c.Flags().StringVar(&created, "created", "", "set created date (YYYY-MM-DD or RFC3339)")
	c.Flags().StringVar(&updated, "updated", "", "set updated date (YYYY-MM-DD or RFC3339)")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newDocMoveCmd() *cobra.Command {
	var newCode string
	var asJSON bool
	c := &cobra.Command{
		Use:   "move DOC-CODE --code NEWCODE",
		Short: "Rename (recode) a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			d, err := s.RecodeDoc(args[0], newCode)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, d)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Recoded to %s\n", d.Code)
			return nil
		},
	}
	c.Flags().StringVar(&newCode, "code", "", "new code (required)")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	c.MarkFlagRequired("code")
	return c
}

func newDocUnarchiveCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "unarchive DOC-CODE",
		Short: "Restore an archived document to active",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.UnarchiveDoc(args[0]); err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, map[string]any{"unarchived": true, "code": args[0]})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Unarchived %s\n", args[0])
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newDocArchiveCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "archive DOC-CODE",
		Short: "Archive a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.ArchiveDoc(args[0]); err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, map[string]any{"archived": true, "code": args[0]})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Archived %s\n", args[0])
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newDocRmCmd() *cobra.Command {
	var force, asJSON bool
	c := &cobra.Command{
		Use:   "rm DOC-CODE",
		Short: "Delete a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				return fmt.Errorf("refusing to delete %s without --force", args[0])
			}
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.DeleteDoc(args[0]); err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, map[string]any{"deleted": true, "code": args[0]})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s\n", args[0])
			return nil
		},
	}
	c.Flags().BoolVar(&force, "force", false, "confirm deletion")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newDocLinkCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "link DOC-CODE T-CODE",
		Short: "Link a document to a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.Link(args[0], args[1]); err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, map[string]any{"linked": true, "doc": args[0], "task": args[1]})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Linked %s <-> %s\n", args[0], args[1])
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newDocUnlinkCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "unlink DOC-CODE T-CODE",
		Short: "Unlink a document from a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.Unlink(args[0], args[1]); err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, map[string]any{"unlinked": true, "doc": args[0], "task": args[1]})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Unlinked %s <-> %s\n", args[0], args[1])
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}
