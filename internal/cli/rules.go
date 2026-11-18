package cli

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/output"
	"github.com/infrashift/lexis/internal/ruleset"
)

func newRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Show the rules of a ruleset",
		Args:  cobra.NoArgs,
	}
	lf := &rulesetFlags{}
	list := &cobra.Command{
		Use:   "list",
		Short: "List the rules, their status and their word limits after each change",
		Args:  cobra.NoArgs,
		RunE:  lf.runList, // nasa-lint: allow(R9) cobra calls the command through RunE.
	}
	lf.register(list.Flags())
	inspect := &cobra.Command{
		Use:   "inspect <ruleset>",
		Short: "Show each rule, its parameters, the lexicon and the vocabulary of a ruleset",
		Args:  cobra.ExactArgs(1),
		RunE:  runRulesInspect, // nasa-lint: allow(R9) cobra calls the command through RunE.
	}
	cmd.AddCommand(list, inspect)
	assert.True(len(cmd.Commands()) == 2, "rules has two subcommands")
	assert.True(list.Flags().Lookup("ruleset") != nil, "rules list takes --ruleset")
	return cmd
}

func (f *rulesetFlags) runList(cmd *cobra.Command, args []string) error {
	assert.True(cmd != nil, "cobra gives the command")
	assert.True(len(args) == 0, "cobra.NoArgs rejects arguments")
	opts, err := f.options()
	if err != nil {
		return err
	}
	rs, err := ruleset.Load(opts)
	if err != nil {
		return err
	}
	out := output.New(cmd.OutOrStdout())
	all := rs.Sorted()
	out.Printf("ruleset %s: %d rules, %d active, %d omitted\n", rs.Name, len(all), len(rs.Active()), len(all)-len(rs.Active()))
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 4, 2, ' ', 0)
	table := output.New(tw)
	table.Printf("ID\tSTATUS\tSEVERITY\tCHECK\tLIMIT\tTITLE\n")
	for _, r := range all {
		table.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", r.ID, status(r), r.Severity, r.Check, limit(r), r.Title)
	}
	if err := table.Err(); err != nil {
		return err
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	out.Printf("%s", buf.String())
	return out.Err()
}

func status(r ruleset.Rule) string {
	assert.True(r.ID != "", "the rule has an ID")
	assert.True(r.Check != "", "the rule has a check")
	if r.Enabled {
		return "active"
	}
	return "omitted"
}

// limit returns the word limit of a max_words rule, or "-".
func limit(r ruleset.Rule) string {
	assert.True(r.ID != "", "the rule has an ID")
	assert.True(r.Params.MaxWords >= 0, "a word limit is not negative")
	if r.Check != "max_words" {
		return "-"
	}
	return fmt.Sprint(r.Params.MaxWords)
}

func runRulesInspect(cmd *cobra.Command, args []string) error {
	assert.True(cmd != nil, "cobra gives the command")
	assert.True(len(args) == 1, "cobra.ExactArgs(1) gives one argument")
	rs, err := ruleset.Load(ruleset.Options{Rulesets: args})
	if err != nil {
		return err
	}
	out := output.New(cmd.OutOrStdout())
	out.Printf("ruleset %s: %s\n", rs.Name, rs.Description)
	for _, r := range rs.Sorted() {
		out.Printf("\n%s (%s, %s)\n  title:   %s\n  scope:   %s\n  check:   %s\n  message: %s\n",
			r.ID, r.Severity, status(r), r.Title, r.Scope, r.Check, r.Message)
		for _, line := range paramLines(r.Params) {
			out.Printf("  %s\n", line)
		}
	}
	writeLexicon(out, rs)
	return out.Err()
}

// writeLexicon writes the lexicon and the vocabulary in sorted order.
func writeLexicon(out *output.Writer, rs *ruleset.Ruleset) {
	assert.True(out != nil, "the writer exists")
	assert.True(rs != nil, "the ruleset exists")
	out.Printf("\nlexicon: %d concepts\n", len(rs.Lexicon))
	for _, name := range slices.Sorted(maps.Keys(rs.Lexicon)) {
		c := rs.Lexicon[name]
		out.Printf("  %s: not %s", c.Approved, strings.Join(c.Unapproved, ", "))
		if len(c.Nominalizations) > 0 {
			out.Printf("; nouns %s", strings.Join(c.Nominalizations, ", "))
		}
		out.Printf("\n")
	}
	out.Printf("\nvocabulary: %d terms\n", len(rs.Vocabulary))
	if len(rs.Vocabulary) > 0 {
		out.Printf("  %s\n", strings.Join(slices.Sorted(maps.Keys(rs.Vocabulary)), ", "))
	}
}

// paramLines formats the parameters that a rule sets, one on each line.
func paramLines(p ruleset.Params) []string {
	assert.True(p.MaxWords >= 0, "a word limit is not negative")
	sets := []struct {
		name  string
		terms map[string]bool
	}{
		{"tokens", p.Tokens}, {"exceptions", p.Exceptions}, {"suffixes", p.Suffixes},
		{"auxiliaries", p.Auxiliaries}, {"participles", p.Participles}, {"verbs", p.Verbs}, {"joiners", p.Joiners},
	}
	var lines []string
	if p.MaxWords > 0 {
		lines = append(lines, fmt.Sprintf("max_words: %d", p.MaxWords))
	}
	if len(p.ProceduralMarkers) > 0 {
		lines = append(lines, "procedural_markers: "+strings.Join(p.ProceduralMarkers, ", "))
	}
	if p.Pattern != "" {
		lines = append(lines, "pattern: "+p.Pattern)
	}
	if len(p.Swap) > 0 {
		pairs := make([]string, 0, len(p.Swap))
		for _, k := range slices.Sorted(maps.Keys(p.Swap)) {
			pairs = append(pairs, k+" -> "+p.Swap[k])
		}
		lines = append(lines, "swap: "+strings.Join(pairs, ", "))
	}
	for _, s := range sets {
		if len(s.terms) > 0 {
			lines = append(lines, s.name+": "+strings.Join(slices.Sorted(maps.Keys(s.terms)), ", "))
		}
	}
	assert.True(len(lines) <= 11, "a rule has eleven parameters or fewer")
	return lines
}
