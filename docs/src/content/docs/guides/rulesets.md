---
title: Write a Ruleset
description: Change the ste-100 ruleset with a CUE file, or write a ruleset of your own team.
sidebar:
  order: 2
---

A ruleset is CUE data. lexis unifies the schema, the built-in ruleset and each file that `--ruleset` names. The [ruleset schema](../../reference/ruleset-schema/) lists each field.

## How changes combine

CUE unification does not depend on the order of the files. A field can get one concrete value only. The built-in ruleset writes each value that a user can change as a CUE default, for example `max_words: *20 | int`. A user file replaces a default. Two files that give two different values for one field fail with the path of the field and the position in each file.

## Change the built-in ruleset

1. Write a file, for example `house.cue`.
2. Add the changes below that the project needs.
3. Run `lexis rules list -r ste-100,house.cue` to see the result.

Turn off a rule:

```cue
rules: "STE-100-07": enabled: false
```

Change a word limit:

```cue
rules: "STE-100-01": params: max_words: 15
```

Change a severity:

```cue
rules: "STE-100-05": severity: "error"
```

Add approved technical terms. STE-100-06 and STE-100-07 do not report them:

```cue
vocabulary: {okfw: true, zmake: true}
```

Add a concept to the lexicon. STE-100-05 then reports each unapproved term, and STE-100-06 reports each noun form:

```cue
lexicon: tell2: {approved: "tell", unapproved: ["apprise"], nominalizations: ["apprisal"]}
```

A user file can add a concept, but it can not change the lists of a built-in concept. CUE lists do not merge.

## Keep a vocabulary file

A project can keep its technical terms in one file, for example `vocabulary.cue`, and load it with each scan. The lexis repository does this for its self-check:

```cue
vocabulary: {
	unification: true // CUE unification
	trailing:    true // a trailing comment
}
```

```sh
bin/lexis scan --ruleset ste-100 --ruleset vocabulary.cue .
```

Add a term only when it names a concept of the project, a tool or a standard. Rewrite other text. The vocabulary applies to STE-100-06 and STE-100-07 only. The load fails when a vocabulary term is also an unapproved term of the lexicon or a banned term, for example of STE-100-04 or STE-100-09.

## Add a rule

A new rule needs each field of `#Rule`. The ID has capital letters, digits and hyphens:

```cue
rules: "DOC-CUSTOM-01": {
	severity: "error"
	scope:    "any_sentence"
	check:    "regex_ban"
	title:    "Do not write TODO in documentation"
	message:  "Remove \"{term}\"."
	params: pattern: "\\bTODO\\b"
}
```

## Write a ruleset without ste-100

A file or a directory alone is a complete ruleset. It needs a `name` and its rules:

```cue
name: "house-style"
rules: "HS-01": {
	severity: "warning"
	scope:    "heading"
	check:    "max_words"
	title:    "Limit the length of a heading"
	message:  "The heading has {words} words."
	params: max_words: 8
}
```

Run `lexis rules list -r house/` to load each `.cue` file of the directory `house/`.

## Errors that the schema finds

- A parameter name that the schema does not know: `max_wrods: field not allowed`.
- A check without its required parameter, for example a `regex_ban` rule without `pattern`.
- A term in two roles: approved, and also unapproved or banned. The error names the term, for example `conflict.just: ... approved and also banned by STE-100-04`.
- A rule ID with lowercase letters or spaces.
