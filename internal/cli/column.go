package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/statikowsky/mar/internal/store"
)

func newColumnCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "column", Aliases: []string{"c"}, Short: "Manage board columns"}

	var addAfter, addBefore string
	var addJSON bool
	add := &cobra.Command{
		Use:   "add NAME",
		Short: "Add a column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if addAfter != "" && addBefore != "" {
				return fmt.Errorf("use only one of --after / --before")
			}
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			var col store.Column
			if addBefore != "" {
				col, err = s.AddColumnBefore(args[0], addBefore)
			} else {
				col, err = s.AddColumn(args[0], addAfter)
			}
			if err != nil {
				return err
			}
			if addJSON {
				return printJSON(cmd, col)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added column %q\n", args[0])
			return nil
		},
	}
	add.Flags().StringVar(&addAfter, "after", "", "insert after this column")
	add.Flags().StringVar(&addBefore, "before", "", "insert before this column")
	add.Flags().BoolVar(&addJSON, "json", false, "output JSON")

	var moveAfter, moveBefore string
	var moveJSON bool
	move := &cobra.Command{
		Use:   "move NAME (--before X | --after X)",
		Short: "Reorder a column relative to another",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if (moveBefore == "") == (moveAfter == "") {
				return fmt.Errorf("specify exactly one of --before / --after")
			}
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			target, before := moveAfter, false
			if moveBefore != "" {
				target, before = moveBefore, true
			}
			if err := s.MoveColumn(args[0], target, before); err != nil {
				return err
			}
			if moveJSON {
				return printJSON(cmd, map[string]any{"moved": true, "column": args[0]})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Moved column %q\n", args[0])
			return nil
		},
	}
	move.Flags().StringVar(&moveBefore, "before", "", "move before this column")
	move.Flags().StringVar(&moveAfter, "after", "", "move after this column")
	move.Flags().BoolVar(&moveJSON, "json", false, "output JSON")

	var renameJSON bool
	rename := &cobra.Command{
		Use:   "rename OLD NEW",
		Short: "Rename a column",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.RenameColumn(args[0], args[1]); err != nil {
				return err
			}
			if renameJSON {
				return printJSON(cmd, map[string]any{"renamed": true, "from": args[0], "to": args[1]})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Renamed %q to %q\n", args[0], args[1])
			return nil
		},
	}
	rename.Flags().BoolVar(&renameJSON, "json", false, "output JSON")

	var force, rmJSON bool
	rm := &cobra.Command{
		Use:   "rm NAME",
		Short: "Remove a column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.RemoveColumn(args[0], force); err != nil {
				return err
			}
			if rmJSON {
				return printJSON(cmd, map[string]any{"removed": true, "column": args[0]})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed column %q\n", args[0])
			return nil
		},
	}
	rm.Flags().BoolVar(&force, "force", false, "delete even if it has tasks")
	rm.Flags().BoolVar(&rmJSON, "json", false, "output JSON")

	cmd.AddCommand(add, move, rename, rm)
	return cmd
}
