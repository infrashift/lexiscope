rules: "HS-02": {
	severity: "note"
	scope:    "token"
	check:    "banned_terms"
	title:    "Do not write etc."
	message:  "Do not use \"{term}\"."
	params: tokens: {"etc.": true}
}
