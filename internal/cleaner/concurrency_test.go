package cleaner

import (
	"sync"
	"testing"
)

// Reproduces the pack --clean scenario: many concurrent Clean calls.
// tree-sitter via CGO may be unsafe under concurrent parsing —
// this test catches result corruption.
func TestConcurrentCleanStable(t *testing.T) {
	c := New([]string{"//!"})
	const in = "export const B = () => <div>{/* jsx */}</div>; // tail\n"
	want := c.Clean("b.tsx", in, true).Content // single-thread reference

	var wg sync.WaitGroup
	errs := make(chan string, 200)
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got := c.Clean("b.tsx", in, true).Content
			if got != want {
				errs <- got
			}
		}()
	}
	wg.Wait()
	close(errs)
	if len(errs) > 0 {
		t.Fatalf("concurrent cleaning is unstable: %d discrepancies, example: %q (reference %q)", len(errs), <-errs, want)
	}
	t.Logf("reference: %q", want)
}
