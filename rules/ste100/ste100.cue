// The ste-100 ruleset: a subset of ASD-STE100 (ADR-0005).
//
// STE-100-01 to STE-100-04 form the core group. They copy the four checks of
// scripts/ste-check.py exactly. STE-100-05 to STE-100-11 form the extended
// group. Each extended rule gives a warning by default.
//
// Each value that a user can change is a default (*value). A user file or a
// --set flag replaces it. A user file adds terms to a set with `term: true`.

name:        *"ste-100" | string
description: *"A subset of ASD-STE100 Simplified Technical English" | string

// One approved term for each concept. STE-100-05 and STE-100-06 come from
// this map, so a user file that adds a concept changes both rules.
lexicon: {
	use: {approved: "use", unapproved: ["utilize", "leverage", "employ"], nominalizations: ["usage", "utilization"]}
	verify: {approved: "verify", unapproved: ["validate", "confirm", "ensure"], nominalizations: ["verification", "validation", "confirmation"]}
	start: {approved: "start", unapproved: ["commence", "initiate"], nominalizations: ["commencement", "initiation"]}
	change: {approved: "change", unapproved: ["modify", "alter", "amend"], nominalizations: ["modification", "alteration", "amendment"]}
	stop: {approved: "stop", unapproved: ["terminate", "cease", "halt"], nominalizations: ["termination", "cessation"]}
	help: {approved: "help", unapproved: ["assist", "aid", "facilitate"], nominalizations: ["assistance", "facilitation"]}
	get: {approved: "get", unapproved: ["obtain", "acquire", "procure"], nominalizations: ["acquisition", "procurement"]}
	show: {approved: "show", unapproved: ["indicate", "demonstrate"], nominalizations: ["indication", "demonstration"]}
	tell: {approved: "tell", unapproved: ["inform", "notify"], nominalizations: ["notification"]}
	give: {approved: "give", unapproved: ["provide", "supply"], nominalizations: ["provision"]}
	remove: {approved: "remove", unapproved: ["eliminate"], nominalizations: ["elimination", "removal"]}
	keep: {approved: "keep", unapproved: ["retain"], nominalizations: ["retention"]}
	find: {approved: "find", unapproved: ["locate"], nominalizations: []}
	"try": {approved: "try", unapproved: ["attempt", "endeavor"], nominalizations: []}
	about: {approved: "about", unapproved: ["approximately"], nominalizations: []}
	enough: {approved: "enough", unapproved: ["sufficient", "adequate"], nominalizations: []}
	so: {approved: "so", unapproved: ["therefore", "thus", "hence"], nominalizations: []}
	because: {approved: "because", unapproved: ["due to"], nominalizations: []}
	before: {approved: "before", unapproved: ["prior to"], nominalizations: []}
	after: {approved: "after", unapproved: ["subsequent to"], nominalizations: []}
}

vocabulary: {}

rules: {
	"STE-100-01": {
		severity: *"error" | #Severity
		scope:    "procedural_sentence"
		check:    "max_words"
		title:    "Limit the length of a procedural step"
		message:  "The step has {words} words. Use {max_words} words or fewer."
		params: {
			max_words: *20 | int
			// The oracle treats only numbered list items as steps.
			procedural_markers: *["numbered"] | [...("numbered" | "bullet")]
		}
	}
	"STE-100-02": {
		severity: *"error" | #Severity
		scope:    "descriptive_sentence"
		check:    "max_words"
		title:    "Limit the length of a descriptive sentence"
		message:  "The sentence has {words} words. Use {max_words} words or fewer."
		params: max_words: *25 | int
	}
	"STE-100-03": {
		severity: *"warning" | #Severity
		scope:    "any_sentence"
		check:    "passive_voice"
		title:    "Use the active voice"
		message:  "\"{term}\" can be passive voice. Name the actor and use the active voice."
		params: {
			auxiliaries: {is: true, are: true, was: true, were: true, be: true, been: true, being: true}
			suffixes: {ed: true}
		}
	}
	"STE-100-04": {
		severity: *"error" | #Severity
		scope:    "any_sentence"
		check:    "banned_terms"
		title:    "Do not use the banned words and phrases"
		message:  "Do not use \"{term}\"."
		params: tokens: {
			should: true, will: true, "won't": true, would: true, please: true
			simply: true, just: true, very: true, really: true, easily: true
			"is expected to": true, "are expected to": true, "can be seen": true
		}
	}
	"STE-100-05": {
		severity: *"warning" | #Severity
		scope:    "any_sentence"
		check:    "term_substitution"
		title:    "Use the approved term for each concept"
		message:  "Use \"{replacement}\", not \"{term}\"."
		params: swap: {for _, c in lexicon for _, u in c.unapproved {(u): c.approved}}
	}
	"STE-100-06": {
		severity: *"warning" | #Severity
		scope:    "any_sentence"
		check:    "nominalizations"
		title:    "Use a verb, not a noun form of the verb"
		message:  "\"{term}\" is a noun form of a verb. Use the verb {replacement}."
		params: {
			suffixes: {tion: true, sion: true, ment: true, ance: true, ence: true}
			swap: {for _, c in lexicon for _, n in c.nominalizations {(n): c.approved}}
			// Technical nouns that end in a suffix but name a thing.
			exceptions: {
				action: true, annotation: true, application: true, assertion: true, caution: true
				collection: true, condition: true, configuration: true, connection: true, declaration: true
				definition: true, description: true, direction: true, documentation: true, duration: true
				exception: true, expression: true, extension: true, function: true, information: true
				instruction: true, iteration: true, location: true, mention: true, notation: true
				operation: true, option: true, permission: true, portion: true, position: true
				question: true, recursion: true, relation: true, section: true, selection: true
				session: true, solution: true, version: true, dimension: true, conversion: true
				argument: true, assignment: true, comment: true, document: true, element: true
				environment: true, fragment: true, increment: true, segment: true, statement: true
				balance: true, distance: true, instance: true, tolerance: true, maintenance: true
				consequence: true, difference: true, evidence: true, occurrence: true, preference: true
				reference: true, sentence: true, sequence: true, licence: true
			}
		}
	}
	"STE-100-07": {
		severity: *"warning" | #Severity
		scope:    "any_sentence"
		check:    "ambiguous_gerunds"
		title:    "Avoid unclear -ing forms"
		message:  "\"{term}\" is an -ing form. Use a verb or a noun."
		params: exceptions: {
			anything: true, bring: true, building: true, ceiling: true, during: true
			encoding: true, everything: true, heading: true, king: true, listing: true
			logging: true, mapping: true, morning: true, nothing: true, padding: true
			ring: true, setting: true, sibling: true, something: true, spring: true
			string: true, thing: true, warning: true, wing: true, binding: true
		}
	}
	"STE-100-08": {
		severity: *"warning" | #Severity
		scope:    "any_sentence"
		check:    "possessive_pronouns"
		title:    "Do not use \"your\""
		message:  "Do not use \"{term}\". Use \"the\"."
		params: tokens: {your: true, yours: true}
	}
	"STE-100-09": {
		severity: *"warning" | #Severity
		scope:    "any_sentence"
		check:    "banned_terms"
		title:    "Do not use marketing words"
		message:  "\"{term}\" is a marketing word. State a fact."
		params: tokens: {
			seamless: true, seamlessly: true, "blazingly fast": true, "blazing fast": true
			"lightning-fast": true, robust: true, effortless: true, effortlessly: true
			powerful: true, modern: true, "out-of-the-box": true, "out of the box": true
			"cutting-edge": true, "state-of-the-art": true, "best-in-class": true
			"world-class": true, "next-generation": true, revolutionary: true
			"game-changing": true, intuitive: true, awesome: true, amazing: true
		}
	}
	"STE-100-10": {
		severity: *"warning" | #Severity
		scope:    "any_sentence"
		check:    "passive_voice"
		title:    "Use the active voice with irregular verbs"
		message:  "\"{term}\" can be passive voice. Name the actor and use the active voice."
		params: {
			auxiliaries: {am: true, is: true, are: true, was: true, were: true, be: true, been: true, being: true}
			participles: {
				beaten: true, begun: true, bound: true, broken: true, brought: true, built: true
				bought: true, caught: true, chosen: true, cut: true, done: true, drawn: true
				driven: true, eaten: true, fallen: true, fed: true, felt: true, forbidden: true
				forgotten: true, found: true, frozen: true, given: true, gotten: true, grown: true
				held: true, hidden: true, hit: true, hung: true, hurt: true, kept: true
				known: true, laid: true, led: true, left: true, lost: true, made: true
				meant: true, overridden: true, paid: true, proven: true, put: true, read: true
				rewritten: true, run: true, said: true, seen: true, sent: true, set: true
				shown: true, shut: true, sold: true, split: true, spoken: true, spread: true
				stolen: true, struck: true, taken: true, taught: true, thought: true, thrown: true
				told: true, torn: true, understood: true, withdrawn: true, won: true, worn: true
				written: true
			}
		}
	}
	"STE-100-11": {
		severity: *"warning" | #Severity
		scope:    "procedural_sentence"
		check:    "one_instruction_per_sentence"
		title:    "Write one instruction in each sentence"
		message:  "The sentence has more than one instruction. Write one instruction in each sentence."
		params: {
			joiners: {and: true, then: true}
			verbs: {
				add: true, apply: true, build: true, change: true, check: true, click: true
				close: true, connect: true, copy: true, create: true, delete: true, edit: true
				enter: true, find: true, get: true, give: true, go: true, install: true
				keep: true, look: true, make: true, move: true, open: true, pull: true
				push: true, put: true, read: true, remove: true, replace: true, run: true
				save: true, select: true, set: true, show: true, start: true, stop: true
				tell: true, test: true, turn: true, type: true, update: true, use: true
				verify: true, wait: true, write: true
			}
		}
	}
}
