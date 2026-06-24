package sortx

import (
	"reflect"
	"testing"
)

func TestLocaleStringsASCII(t *testing.T) {
	got := []string{"banana", "apple", "cherry"}
	LocaleStrings(got)
	want := []string{"apple", "banana", "cherry"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LocaleStrings = %v, want %v", got, want)
	}
}

// Diacritics sort next to their base letter (ICU root collation), not after "z".
func TestLocaleStringsDiacritics(t *testing.T) {
	got := []string{"z", "ä", "a"}
	LocaleStrings(got)
	want := []string{"a", "ä", "z"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LocaleStrings = %v, want %v", got, want)
	}
}

// Sorting happens in place: the same backing slice is reordered.
func TestLocaleStringsInPlace(t *testing.T) {
	s := []string{"c", "a", "b"}
	LocaleStrings(s)
	if s[0] != "a" || s[1] != "b" || s[2] != "c" {
		t.Fatalf("expected in-place sort, got %v", s)
	}
}

func TestLocaleStringsEdgeCases(t *testing.T) {
	var empty []string
	LocaleStrings(empty) // must not panic

	single := []string{"only"}
	LocaleStrings(single)
	if len(single) != 1 || single[0] != "only" {
		t.Fatalf("single-element slice changed: %v", single)
	}
}
