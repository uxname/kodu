package registry

import (
	"os"
	"path/filepath"
	"testing"
)

// isolate points the registry at a temporary directory via XDG_CONFIG_HOME.
func isolate(t *testing.T) *Service {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return New()
}

func TestLoadEmptyWhenNoFile(t *testing.T) {
	s := isolate(t)
	reg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Projects) != 0 {
		t.Fatalf("expected empty registry, got %v", reg.Projects)
	}
}

func TestAddGetRemove(t *testing.T) {
	s := isolate(t)
	if err := s.Add("demo", ProjectEntry{Path: "/x", Stands: DefaultStands()}, false); err != nil {
		t.Fatal(err)
	}
	e, ok, err := s.Get("demo")
	if err != nil || !ok {
		t.Fatalf("Get demo: ok=%v err=%v", ok, err)
	}
	if e.Path != "/x" || len(e.Stands) != 4 {
		t.Fatalf("entry = %+v", e)
	}
	if err := s.Remove("demo"); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := s.Get("demo"); ok {
		t.Fatal("project should have been removed")
	}
}

func TestAddDuplicateRejected(t *testing.T) {
	s := isolate(t)
	_ = s.Add("demo", ProjectEntry{Path: "/x", Stands: DefaultStands()}, false)
	if err := s.Add("demo", ProjectEntry{Path: "/y", Stands: DefaultStands()}, false); err == nil {
		t.Fatal("duplicate name without overwrite should return an error")
	}
	if err := s.Add("demo", ProjectEntry{Path: "/y", Stands: DefaultStands()}, true); err != nil {
		t.Fatalf("overwrite should succeed: %v", err)
	}
}

func TestRemoveMissing(t *testing.T) {
	s := isolate(t)
	if err := s.Remove("nope"); err == nil {
		t.Fatal("removing a missing project should return an error")
	}
}

func TestSaveAtomicAndStandsDefault(t *testing.T) {
	s := isolate(t)
	// Writing without stands -> the default is applied on load.
	if err := s.Add("demo", ProjectEntry{Path: "/x"}, false); err != nil {
		t.Fatal(err)
	}
	reg, _ := s.Load()
	if len(reg.Projects["demo"].Stands) != 4 {
		t.Fatalf("stands default was not applied: %v", reg.Projects["demo"].Stands)
	}
	// The file is actually created and contains a trailing newline.
	data, err := os.ReadFile(s.FilePath())
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Fatal("expected a trailing newline")
	}
	if filepath.Base(s.FilePath()) != "registry.json" {
		t.Fatalf("unexpected file name: %s", s.FilePath())
	}
}
