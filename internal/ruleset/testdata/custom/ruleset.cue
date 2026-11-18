name: "house-style"
rules: "HS-01": {
	severity: "warning"
	scope:    "heading"
	check:    "max_words"
	title:    "A heading has 8 words or fewer"
	message:  "The heading has {words} words."
	params: max_words: 8
}
