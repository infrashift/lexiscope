// Adds a rule to ste-100.
rules: "DOC-CUSTOM-01": {
	severity: "error"
	scope:    "any_sentence"
	check:    "regex_ban"
	title:    "Do not write TODO in documentation"
	message:  "Remove \"{term}\"."
	params: pattern: "\\bTODO\\b"
}
