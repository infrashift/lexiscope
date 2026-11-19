// The semscope overlay of lexis. Use it with the nasa-p10 ruleset:
//   semscope scan -r nasa-p10 -r .semscope.cue .
// ADR-0003 adapts the Power of 10 rules to Go. This file applies the same
// changes to nasa-p10, so semscope and cmd/nasa-lint give the same findings.
package overlay

// R3 deviation (ADR-0003): Go allocates on the heap. lexis sets a hard cap
// on each buffer and allocates each buffer once per file. Review checks R3.
rules: "NASA-P10-03": enabled: false

// R6 (ADR-0003): declare each variable in the smallest block. The compiler
// and review check R6. lexis keeps package variables for fixed tables,
// sentinel errors, embedded CUE sources and compiled regular expressions.
rules: "NASA-P10-06": enabled: false

// The Python scripts are tooling from zmake, not lexis code. lexis does not
// change them (CLAUDE.md).
exclude: scripts: "Python lint tooling from zmake; lexis does not change it"
