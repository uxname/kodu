package tokenizer

import (
	"math"
	"testing"
)

// Reference values captured from js-tiktoken o200k_base (parity with the Node version).
func TestCountParity(t *testing.T) {
	tk := New()
	cases := []struct {
		text string
		want int
	}{
		{"hello world", 2},
		{"", 0},
		{"const x = 42;\n", 6},
		{"Привет, мир! 🌍", 7},
	}
	for _, c := range cases {
		est, err := tk.Count(c.text)
		if err != nil {
			t.Fatalf("Count(%q) error: %v", c.text, err)
		}
		if est.Tokens != c.want {
			t.Fatalf("Count(%q) = %d tokens, wanted %d", c.text, est.Tokens, c.want)
		}
	}
}

func TestUSDEstimate(t *testing.T) {
	tk := New()
	est, err := tk.Count("hello world") // 2 tokens
	if err != nil {
		t.Fatal(err)
	}
	want := 2.0 / 1_000_000 * DefaultPricePerMillion
	if math.Abs(est.USDEstimate-want) > 1e-12 {
		t.Fatalf("usd = %v, wanted %v", est.USDEstimate, want)
	}
}
