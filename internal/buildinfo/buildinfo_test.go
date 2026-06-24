package buildinfo

import "testing"

// The "from source" defaults are used when the linker does not override them.
func TestDefaults(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"Version", Version, "dev"},
		{"Commit", Commit, "none"},
		{"Date", Date, "unknown"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}
