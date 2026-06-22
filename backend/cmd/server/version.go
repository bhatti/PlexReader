// SPDX-License-Identifier: LGPL-2.1-or-later
package main

// Version is the semantic version string, injected at build time via -ldflags.
// Format: MAJOR.MINOR.PATCH-GITHASH (e.g. 0.1.1-abc1234)
// Falls back to "dev" when running with `go run`.
var Version = "dev"
