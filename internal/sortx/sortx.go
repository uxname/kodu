// Package sortx сортирует строки так же, как JS String.prototype.localeCompare
// (корневая ICU-коллация) — для паритета порядка файлов с Node-версией.
package sortx

import (
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// LocaleStrings сортирует срез на месте по корневой ICU-коллации.
func LocaleStrings(s []string) {
	collate.New(language.Und).SortStrings(s)
}
