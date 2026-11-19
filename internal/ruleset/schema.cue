// The schema of a lexiscope ruleset (ADR-0004). internal/ruleset unifies
// #Ruleset with the built-in ruleset, each user file and the CLI flags.
// Each ruleset file can refer to the definitions of this file.

#Severity: "error" | "warning" | "note"

#Scope: "procedural_sentence" | "descriptive_sentence" | "any_sentence" | "heading" | "comment" | "token"

#Check: "max_words" | "banned_terms" | "term_substitution" | "passive_voice" |
	"nominalizations" | "ambiguous_gerunds" | "possessive_pronouns" |
	"one_instruction_per_sentence" | "regex_ban"

#RuleID: =~"^[A-Z][A-Z0-9]*(-[A-Z0-9]+)*$"

// A set of terms. A user file adds a term with `term: true`.
#Terms: [string]: true

// The parameters of a rule. The struct is closed, so a misspelled name
// fails. A built-in value that a user can change is a default (*value).
#Params: {
	max_words?: int & >0
	procedural_markers?: [...("numbered" | "bullet")]
	tokens?:      #Terms
	swap?:        [string]: string & !=""
	exceptions?:  #Terms
	suffixes?:    #Terms
	auxiliaries?: #Terms
	participles?: #Terms
	verbs?:       #Terms
	joiners?:     #Terms
	pattern?:     string & !=""
}

#Rule: {
	id:       #RuleID
	enabled:  bool | *true
	severity: #Severity
	scope:    #Scope
	check:    #Check
	title:    string & !=""
	// The engine replaces {name} with a value of the finding, for example
	// {words}, {max_words}, {term} or {replacement}.
	message: string & !=""
	params:  #Params

	// Each check needs these parameters.
	if check == "max_words" {params: max_words: _}
	if check == "banned_terms" || check == "possessive_pronouns" {params: tokens: _}
	if check == "term_substitution" {params: swap: _}
	if check == "passive_voice" {params: auxiliaries: _}
	if check == "nominalizations" {params: suffixes: _}
	if check == "one_instruction_per_sentence" {params: verbs: _}
	if check == "regex_ban" {params: pattern: _}
}

// A concept of the lexicon: one approved term, its unapproved synonyms and
// its noun forms.
#Concept: {
	approved: string & !=""
	unapproved: [...string]
	nominalizations: [...string]
}

#Ruleset: {
	name:        string & !=""
	description?: string
	lexicon: [string]: #Concept
	// Approved technical terms. The extended rules do not report them.
	vocabulary: #Terms
	rules: [ID=string]: #Rule & {id: ID}

	// An approved term can not be unapproved or banned. Each conflict is a
	// field with two values that do not unify, so the load fails and the
	// error names the term.
	_approved: {
		for _, c in lexicon {(c.approved): true}
		vocabulary
	}
	for _, c in lexicon for _, u in c.unapproved if _approved[u] != _|_ {
		conflict: (u): "approved and also unapproved in the lexicon" & false
	}
	for id, r in rules if r.check == "banned_terms" for t, _ in r.params.tokens if _approved[t] != _|_ {
		conflict: (t): "approved and also banned by \(id)" & false
	}
}
