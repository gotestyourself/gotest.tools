/*

Command gty-migration-from-testify migrates one or more packages from
testify/assert and testify/require to gotestyourself/assert.

To run on all packages (including external test packages) use:

	go list \
		-f '{{.ImportPath}} {{if .XTestGoFiles}}{{"\n"}}{{.ImportPath}}_test{{end}}' \
		./... | xargs gty-migrate-from-testify

The cmp package can be aliases to make the assertions more readable:

    gty-migrate-from-testify --import-cmp-alias=is

*/

package main
