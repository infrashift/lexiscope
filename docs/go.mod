// This file keeps the docs site out of the lexiscope Go module. Without it,
// `go list ./...` also walks docs/node_modules, and a .go file in an npm
// package would join the vet and the tests of lexiscope. ADR-0007 records it.
module github.com/infrashift/lexiscope/docs

go 1.26
