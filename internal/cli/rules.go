package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/infrashift/lexis/internal/assert"
)

func newRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Show the rules of a ruleset",
		Args:  cobra.NoArgs,
	}
	list := &cobra.Command{
		Use:   "list",
		Short: "List the active rules, the omitted rules and the word limits",
		Args:  cobra.NoArgs,
		RunE:  runRulesList, // nasa-lint: allow(R9) cobra calls the command through RunE.
	}
	list.Flags().StringSliceP("ruleset", "r", []string{"ste-100"}, "built-in ruleset name, or CUE file or directory")
	inspect := &cobra.Command{
		Use:   "inspect <ruleset>",
		Short: "Show each rule, the lexicon and the word limits of a ruleset",
		Args:  cobra.ExactArgs(1),
		RunE:  runRulesInspect, // nasa-lint: allow(R9) cobra calls the command through RunE.
	}
	cmd.AddCommand(list, inspect)
	assert.True(len(cmd.Commands()) == 2, "rules has two subcommands")
	assert.True(list.Flags().Lookup("ruleset") != nil, "rules list takes --ruleset")
	return cmd
}

func runRulesList(cmd *cobra.Command, args []string) error {
	assert.True(cmd != nil, "cobra gives the command")
	assert.True(len(args) == 0, "cobra.NoArgs rejects arguments")
	return fmt.Errorf("rules list: %w: the ruleset engine arrives in Phase 3 of PLAN.md", errNotAvailable)
}

func runRulesInspect(cmd *cobra.Command, args []string) error {
	assert.True(cmd != nil, "cobra gives the command")
	assert.True(len(args) == 1, "cobra.ExactArgs(1) gives one argument")
	return fmt.Errorf("rules inspect %s: %w: the ruleset engine arrives in Phase 3 of PLAN.md", args[0], errNotAvailable)
}
