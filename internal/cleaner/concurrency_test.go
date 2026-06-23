package cleaner

import (
	"sync"
	"testing"
)

// Воспроизводит сценарий pack --clean: много параллельных Clean.
// tree-sitter через CGO может быть небезопасен при конкурентном парсинге —
// этот тест ловит порчу результата.
func TestConcurrentCleanStable(t *testing.T) {
	c := New([]string{"//!"})
	const in = "export const B = () => <div>{/* jsx */}</div>; // tail\n"
	want := c.Clean("b.tsx", in, true).Content // эталон single-thread

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
		t.Fatalf("конкурентная очистка нестабильна: %d расхождений, пример: %q (эталон %q)", len(errs), <-errs, want)
	}
	t.Logf("эталон: %q", want)
}
