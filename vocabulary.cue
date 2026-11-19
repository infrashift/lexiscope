// The vocabulary of the lexiscope repository: technical terms of this project
// that the extended rules STE-100-06 and STE-100-07 do not report. The
// self-check loads this file over the built-in ste-100 ruleset:
//
//	lexiscope scan --ruleset ste-100 --ruleset vocabulary.cue ...
//
// Add a term only when it names a concept of the project, a tool or a
// standard. Rewrite other prose.
vocabulary: {
	// Project process: ADRs, issues and the Power of 10 rules.
	acceptance:  true // acceptance criteria of an issue
	decision:    true // architecture decision record
	deviation:   true // a Power of 10 rule that lexiscope does not follow
	development: true // the development workflow
	resolution:  true // the Resolution section of an issue
	allocation:  true // Power of 10 rule R3
	generation:  true // code generation, Power of 10 rule R8
	indirection: true // Power of 10 rule R9
	suppression: true // a nasa-lint allow comment
	violation:   true // a nasa-lint finding
	mutation:    true // mutation tests

	// CUE and the ruleset.
	comprehension:  true // a CUE comprehension
	unification:    true // CUE unification
	projection:     true // the lexicon projection into rules
	substitution:   true // the term substitution check
	nominalization: true // the nominalizations check
	replacement:    true // the replacement text of a finding
	marketing:      true // the marketing words check

	// Text processing, the CLI and SARIF.
	folding:    true // case folding
	trailing:   true // a trailing comment
	usage:      true // a usage error
	validation: true // SARIF validation
	scanning:   true // code scanning tools
	landing:    true // a landing page
}
