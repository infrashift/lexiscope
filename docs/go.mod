// This file keeps the docs site out of the lexis Go module. Without it,
// `go list ./...` also walks docs/node_modules, and a .go file in an npm
// package would join the vet and the tests of lexis. ADR-0007 records it.
module github.com/infrashift/lexis/docs

go 1.26
