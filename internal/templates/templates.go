// Package templates embeds goforge's scaffold templates and the vendored
// conventions library, so the goforge binary carries everything it needs to
// generate a project with no network access.
package templates

import "embed"

// Base is the static skeleton of every scaffolded project: cmd/ entrypoints,
// build tooling, CI, and the shared internal/ support packages. Files ending
// in .tmpl are Go text/template files; everything else is copied verbatim.
//
//go:embed all:base
var Base embed.FS

// Domain is the generic vertical-slice template used both for the baked-in
// example domain at `goforge new` time and for `goforge generate domain`.
//
//go:embed all:domain
var Domain embed.FS

// Skills is a vendored, point-in-time snapshot of the golang-* conventions
// library (.claude/skills, .agents/rules), copied verbatim into every
// scaffolded project. It is not resynced automatically; see README.md for
// how to pull in updates.
//
//go:embed all:skills
var Skills embed.FS
