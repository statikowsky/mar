package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/statikowsky/mar/internal/store"
)

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "task", Aliases: []string{"t"}, Short: "Manage tasks"}
	cmd.AddCommand(
		newTaskCreateCmd(), newTaskListCmd(), newTaskShowCmd(),
		newTaskEditCmd(), newTaskMoveCmd(), newTaskRmCmd(),
		newTaskArchiveCmd(), newTaskUnarchiveCmd(),
		newTaskLinkCmd(), newTaskUnlinkCmd(),
	)
	return cmd
}

func newTaskCreateCmd() *cobra.Command {
	var title, column, bodyPath, code string
	var asJSON bool
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a task",
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
			tk, err := s.CreateTaskWithCode(code, title, body, column)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, tk)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", tk.Code)
			return nil
		},
	}
	c.Flags().StringVar(&title, "title", "", "task title (required)")
	c.Flags().StringVar(&code, "code", "", "optional task code, e.g. 39 or T-39 (default: auto-numbered)")
	c.Flags().StringVar(&column, "column", "", "column name (default: first column)")
	c.Flags().StringVar(&bodyPath, "body", "", "markdown notes file, or -")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	c.MarkFlagRequired("title")
	return c
}

func newTaskListCmd() *cobra.Command {
	var column, status string
	var asJSON bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			tasks, err := s.ListTasks(column, status)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, tasks)
			}
			for _, t := range tasks {
				fmt.Fprintf(cmd.OutOrStdout(), "%-6s %s\n", t.Code, t.Title)
			}
			return nil
		},
	}
	c.Flags().StringVar(&column, "column", "", "filter by column")
	c.Flags().StringVar(&status, "status", "active", "filter by status: active|archived (empty for all)")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newTaskShowCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "show T-CODE",
		Short: "Show a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			tk, err := s.GetTask(args[0])
			if err != nil {
				return err
			}
			if asJSON {
				docs, err := s.DocCodesForTask(tk.Code)
				if err != nil {
					return err
				}
				return printJSON(cmd, struct {
					store.Task
					Docs []string `json:"docs,omitempty"`
				}{tk, docs})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n\n%s\n", tk.Code, tk.Title, tk.Body)
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newTaskEditCmd() *cobra.Command {
	var title, bodyPath, created, updated string
	var asJSON bool
	c := &cobra.Command{
		Use:   "edit T-CODE",
		Short: "Edit a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			var tp, bp *string
			if cmd.Flags().Changed("title") {
				tp = &title
			}
			if cmd.Flags().Changed("body") {
				body, err := readBody(cmd, bodyPath)
				if err != nil {
					return err
				}
				bp = &body
			}
			tk, err := s.EditTask(args[0], tp, bp)
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
				tk, err = s.SetTaskDates(tk.Code, cp, up)
				if err != nil {
					return err
				}
			}
			if asJSON {
				return printJSON(cmd, tk)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", tk.Code)
			return nil
		},
	}
	c.Flags().StringVar(&title, "title", "", "new title")
	c.Flags().StringVar(&bodyPath, "body", "", "new notes file, or -")
	c.Flags().StringVar(&created, "created", "", "set created date (YYYY-MM-DD or RFC3339)")
	c.Flags().StringVar(&updated, "updated", "", "set updated date (YYYY-MM-DD or RFC3339)")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newTaskMoveCmd() *cobra.Command {
	var column, after string
	var asJSON bool
	c := &cobra.Command{
		Use:   "move T-CODE --column NAME [--after T-CODE]",
		Short: "Move a task to a column and/or reorder it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			var afterPtr *string
			if cmd.Flags().Changed("after") {
				afterPtr = &after
			}
			tk, err := s.MoveTask(args[0], column, afterPtr)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, tk)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Moved %s\n", tk.Code)
			return nil
		},
	}
	c.Flags().StringVar(&column, "column", "", "target column (default: keep current)")
	c.Flags().StringVar(&after, "after", "", "place after this task code")
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newTaskRmCmd() *cobra.Command {
	var force, asJSON bool
	c := &cobra.Command{
		Use:   "rm T-CODE",
		Short: "Delete a task",
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
			if err := s.DeleteTask(args[0]); err != nil {
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

func newTaskArchiveCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "archive T-CODE",
		Short: "Archive a task (hide from the board)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.ArchiveTask(args[0]); err != nil {
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

func newTaskUnarchiveCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "unarchive T-CODE",
		Short: "Restore an archived task to the board",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.UnarchiveTask(args[0]); err != nil {
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

func newBoardCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:     "board",
		Aliases: []string{"b"},
		Short:   "Show the kanban board",
	}
	show := &cobra.Command{
		Use:   "show",
		Short: "Show the board",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			board, err := s.Board()
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, board)
			}
			for _, col := range board {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", col.Name)
				for _, tk := range col.Tasks {
					fmt.Fprintf(cmd.OutOrStdout(), "  %-6s %s\n", tk.Code, tk.Title)
				}
			}
			return nil
		},
	}
	show.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	c.AddCommand(show)
	return c
}

func newTaskLinkCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "link T-CODE DOC-CODE",
		Short: "Link a task to a document",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.Link(args[1], args[0]); err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, map[string]any{"linked": true, "task": args[0], "doc": args[1]})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Linked %s <-> %s\n", args[0], args[1])
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

func newTaskUnlinkCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "unlink T-CODE DOC-CODE",
		Short: "Unlink a task from a document",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.Unlink(args[1], args[0]); err != nil {
				return err
			}
			if asJSON {
				return printJSON(cmd, map[string]any{"unlinked": true, "task": args[0], "doc": args[1]})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Unlinked %s <-> %s\n", args[0], args[1])
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}
