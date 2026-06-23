// Package sortx sorts strings the same way as JS String.prototype.localeCompare
// (the root ICU collation) — for file-order parity with the Node version.
package sortx

import (
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// LocaleStrings sorts the slice in place using the root ICU collation.
func LocaleStrings(s []string) {
	collate.New(language.Und).SortStrings(s)
}
