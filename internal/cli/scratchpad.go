package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/statikowsky/mar/internal/store"
)

func newScratchCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "scratch", Aliases: []string{"scratchpad"}, Short: "Manage scratchpad notes"}
	cmd.AddCommand(newScratchShowCmd(), newScratchAddCmd(), newScratchEditCmd(), newScratchRmCmd(), newScratchPromoteCmd())
	return cmd
}

func newScratchShowCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "show",
		Short: "Show the scratchpad",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			pad, err := s.Scratchpad()
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, pad)
			}
			if len(pad.Notes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Scratchpad is empty.")
				return nil
			}
			for _, note := range pad.Notes {
				line := strings.ReplaceAll(note.Text, "\n", " ")
				fmt.Fprintf(cmd.OutOrStdout(), "%-8s %s\n", note.ID, line)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newScratchAddCmd() *cobra.Command {
	var textPath, color string
	var x, y, width int
	var asJSON bool
	c := &cobra.Command{
		Use:   "add",
		Short: "Add a scratchpad note",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			text, err := readBody(cmd, textPath)
			if err != nil {
				return err
			}
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			note, err := s.CreateScratchNote(text, x, y, width, color)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, note)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", note.ID)
			return nil
		},
	}
	c.Flags().StringVar(&textPath, "text", "", "note text file, or - for stdin (required)")
	c.Flags().IntVar(&x, "x", 0, "horizontal position")
	c.Flags().IntVar(&y, "y", 0, "vertical position")
	c.Flags().IntVar(&width, "width", 0, "note width")
	c.Flags().StringVar(&color, "color", "", "neutral|blue|green|yellow|red|purple")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	c.MarkFlagRequired("text")
	return c
}

func newScratchEditCmd() *cobra.Command {
	var textPath, color, link string
	var x, y, width, z int
	var asJSON bool
	c := &cobra.Command{
		Use:   "edit S-ID",
		Short: "Edit a scratchpad note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			pad, err := s.Scratchpad()
			if err != nil {
				return err
			}
			var note *store.ScratchNote
			for i := range pad.Notes {
				if strings.EqualFold(pad.Notes[i].ID, args[0]) {
					note = &pad.Notes[i]
					break
				}
			}
			if note == nil {
				return fmt.Errorf("scratch note %s: %w", args[0], store.ErrNotFound)
			}
			if cmd.Flags().Changed("text") {
				note.Text, err = readBody(cmd, textPath)
				if err != nil {
					return err
				}
			}
			if cmd.Flags().Changed("x") {
				note.X = x
			}
			if cmd.Flags().Changed("y") {
				note.Y = y
			}
			if cmd.Flags().Changed("width") {
				note.Width = width
			}
			if cmd.Flags().Changed("color") {
				note.Color = color
			}
			if cmd.Flags().Changed("z") {
				note.Z = z
			}
			if cmd.Flags().Changed("link") {
				note.Link = link
			}
			updated, err := s.UpdateScratchNote(*note)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, updated)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", updated.ID)
			return nil
		},
	}
	c.Flags().StringVar(&textPath, "text", "", "new note text file, or - for stdin")
	c.Flags().IntVar(&x, "x", 0, "horizontal position")
	c.Flags().IntVar(&y, "y", 0, "vertical position")
	c.Flags().IntVar(&width, "width", 0, "note width")
	c.Flags().StringVar(&color, "color", "", "note color")
	c.Flags().IntVar(&z, "z", 0, "stacking order")
	c.Flags().StringVar(&link, "link", "", "linked T-* or DOC-* code")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newScratchRmCmd() *cobra.Command {
	var force, asJSON bool
	c := &cobra.Command{
		Use:   "rm S-ID",
		Short: "Delete a scratchpad note",
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
			if err := s.DeleteScratchNote(strings.ToUpper(args[0])); err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, map[string]any{"deleted": true, "code": strings.ToUpper(args[0])})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s\n", strings.ToUpper(args[0]))
			return nil
		},
	}
	c.Flags().BoolVar(&force, "force", false, "confirm deletion")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newScratchPromoteCmd() *cobra.Command {
	var asTask, asDoc, asJSON bool
	var code, docType, column string
	c := &cobra.Command{
		Use:   "promote S-ID",
		Short: "Create a task or document from a scratchpad note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if asTask == asDoc {
				return fmt.Errorf("use exactly one of --task or --doc")
			}
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			pad, err := s.Scratchpad()
			if err != nil {
				return err
			}
			var note *store.ScratchNote
			for i := range pad.Notes {
				if strings.EqualFold(pad.Notes[i].ID, args[0]) {
					note = &pad.Notes[i]
					break
				}
			}
			if note == nil {
				return fmt.Errorf("scratch note %s: %w", args[0], store.ErrNotFound)
			}
			title, body := scratchTitleBody(note.Text)
			var created string
			if asTask {
				task, err := s.CreateTask(title, body, column)
				if err != nil {
					return err
				}
				created = task.Code
			} else {
				if code == "" || docType == "" {
					return fmt.Errorf("--code and --type are required with --doc")
				}
				doc, err := s.CreateDoc(code, title, docType, body)
				if err != nil {
					return err
				}
				created = doc.Code
			}
			note.Link = created
			if _, err := s.UpdateScratchNote(*note); err != nil {
				return fmt.Errorf("created %s but could not link scratch note: %w", created, err)
			}
			if asJSON {
				return printJSON(cmd, map[string]string{"code": created, "note": note.ID})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s from %s\n", created, note.ID)
			return nil
		},
	}
	c.Flags().BoolVar(&asTask, "task", false, "create a task")
	c.Flags().BoolVar(&asDoc, "doc", false, "create a document")
	c.Flags().StringVar(&code, "code", "", "document code (required with --doc)")
	c.Flags().StringVar(&docType, "type", "", "document type (required with --doc)")
	c.Flags().StringVar(&column, "column", "", "task column (default: first column)")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func scratchTitleBody(text string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(text), "\n", 2)
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
