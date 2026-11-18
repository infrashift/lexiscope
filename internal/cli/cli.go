// Package cli holds the cobra commands of lexis and the exit status rules.
//
// Exit status: 0 when the command succeeds without error findings, 1 when a
// scan finds errors (from Phase 4), and 2 on a usage error or a failure.
package cli

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/output"
	"github.com/infrashift/lexis/internal/version"
)

// Exit statuses of lexis.
const (
	ExitOK       = 0
	ExitFindings = 1
	ExitFailure  = 2
)

// maxArgs bounds the command line.
const maxArgs = 1 << 16

// Main runs lexis with args and returns the exit status. It writes each
// failure to stderr.
func Main(args []string, stdout, stderr io.Writer) int {
	assert.True(stdout != nil && stderr != nil, "the output streams exist")
	assert.True(len(args) <= maxArgs, "the command line is bounded")
	code, failure := execute(args, stdout, stderr)
	if failure != "" {
		output.New(stderr).Printf("lexis: %s\n", failure)
	}
	return code
}

// execute is the boundary that turns a failed assertion into exit status 2.
func execute(args []string, stdout, stderr io.Writer) (code int, failure string) {
	assert.True(stdout != nil, "stdout exists")
	assert.True(stderr != nil, "stderr exists")
	defer assert.Recover(&code, &failure)
	root := newRoot(stdout, stderr)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		return ExitFailure, err.Error()
	}
	return ExitOK, ""
}

// newRoot builds the command tree.
func newRoot(stdout, stderr io.Writer) *cobra.Command {
	assert.True(stdout != nil, "stdout exists")
	assert.True(stderr != nil, "stderr exists")
	root := &cobra.Command{
		Use:   "lexis",
		Short: "Check documentation and code comments against CUE rulesets",
		Long: "lexis checks Markdown files and the comments of source files against rulesets in CUE.\n" +
			"The built-in ruleset is ste-100, a subset of ASD-STE100.",
		Version:       version.Get().Details(),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(newVersionCmd(), newScanCmd(), newRulesCmd())
	return root
}

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Write the version line",
		Args:  cobra.NoArgs,
		RunE:  runVersion, // nasa-lint: allow(R9) cobra calls the command through RunE.
	}
	assert.True(cmd.Use == "version", "the command has a name")
	assert.True(cmd.Args != nil, "the command takes no arguments")
	return cmd
}

func runVersion(cmd *cobra.Command, args []string) error {
	assert.True(cmd != nil, "cobra gives the command")
	assert.True(len(args) == 0, "cobra.NoArgs rejects arguments")
	out := output.New(cmd.OutOrStdout())
	out.Printf("%s\n", version.Get().String())
	return out.Err()
}
