---
title: Ruleset Schema
description: The fields of a lexis ruleset, a rule and its parameters.
sidebar:
  order: 3
---

`internal/ruleset/schema.cue` defines the schema. The schema closes each definition, so a field that the schema does not list fails.

## `#Ruleset`

| Field | Type | Meaning |
|---|---|---|
| `name` | string | The name of the ruleset. Required. |
| `description` | string | Optional. |
| `lexicon` | map of `#Concept` | One approved term for each concept. |
| `vocabulary` | set of terms | Approved technical terms. The extended rules do not report them. |
| `rules` | map of `#Rule` | The rules, with the rule ID as the key. |

A set of terms is a CUE struct with `true` values: `{okfw: true}`. A file adds a term to a set. It can not remove a term.

## `#Concept`

| Field | Type | Meaning |
|---|---|---|
| `approved` | string | The approved term. |
| `unapproved` | list of strings | Synonyms. STE-100-05 replaces each with the approved term. |
| `nominalizations` | list of strings | Noun forms. STE-100-06 replaces each with the approved verb. |

## `#Rule`

| Field | Type | Meaning |
|---|---|---|
| `id` | string | The key of the rule. lexis sets it. |
| `enabled` | bool | `true` by default. |
| `severity` | `error`, `warning` or `note` | Required. |
| `scope` | see below | The text that the rule reads. |
| `check` | see below | The check function. |
| `title` | string | A short name of the rule. |
| `message` | string | The text of a finding. The engine replaces `{words}`, `{max_words}`, `{term}` and `{replacement}`. |
| `params` | `#Params` | The parameters of the check. |

Scopes: `procedural_sentence`, `descriptive_sentence`, `any_sentence`, `heading`, `comment` and `token`.

## Checks and their parameters

| Check | Required | Optional |
|---|---|---|
| `max_words` | `max_words` | `procedural_markers` |
| `banned_terms` | `tokens` | `exceptions` |
| `term_substitution` | `swap` | `exceptions` |
| `passive_voice` | `auxiliaries` | `suffixes`, `participles`, `exceptions` |
| `nominalizations` | `suffixes` | `swap`, `exceptions` |
| `ambiguous_gerunds` | none | `exceptions` |
| `possessive_pronouns` | `tokens` | none |
| `one_instruction_per_sentence` | `verbs` | `joiners` |
| `regex_ban` | `pattern` | none |

## `#Params`

| Parameter | Type |
|---|---|
| `max_words` | integer above 0 |
| `procedural_markers` | list of `numbered` and `bullet` |
| `tokens`, `exceptions`, `suffixes`, `auxiliaries`, `participles`, `verbs`, `joiners` | set of terms |
| `swap` | map from a term to its replacement |
| `pattern` | regular expression (RE2 syntax) |

A word check skips code names, for example `skipFenceOrComment`: a word with a capital letter after its first letter, a digit or an underscore.
