package registry

import (
	"os"
	"path/filepath"
	"testing"
)

// isolate направляет реестр в временную директорию через XDG_CONFIG_HOME.
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
		t.Fatalf("ожидал пустой реестр, got %v", reg.Projects)
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
		t.Fatal("проект должен быть удалён")
	}
}

func TestAddDuplicateRejected(t *testing.T) {
	s := isolate(t)
	_ = s.Add("demo", ProjectEntry{Path: "/x", Stands: DefaultStands()}, false)
	if err := s.Add("demo", ProjectEntry{Path: "/y", Stands: DefaultStands()}, false); err == nil {
		t.Fatal("повторное имя без overwrite должно вернуть ошибку")
	}
	if err := s.Add("demo", ProjectEntry{Path: "/y", Stands: DefaultStands()}, true); err != nil {
		t.Fatalf("overwrite должен пройти: %v", err)
	}
}

func TestRemoveMissing(t *testing.T) {
	s := isolate(t)
	if err := s.Remove("nope"); err == nil {
		t.Fatal("удаление отсутствующего должно вернуть ошибку")
	}
}

func TestSaveAtomicAndStandsDefault(t *testing.T) {
	s := isolate(t)
	// Запись без stands → при загрузке проставляется дефолт.
	if err := s.Add("demo", ProjectEntry{Path: "/x"}, false); err != nil {
		t.Fatal(err)
	}
	reg, _ := s.Load()
	if len(reg.Projects["demo"].Stands) != 4 {
		t.Fatalf("stands default не применился: %v", reg.Projects["demo"].Stands)
	}
	// Файл реально создан и содержит trailing newline.
	data, err := os.ReadFile(s.FilePath())
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Fatal("ожидал завершающий перевод строки")
	}
	if filepath.Base(s.FilePath()) != "registry.json" {
		t.Fatalf("неожиданное имя файла: %s", s.FilePath())
	}
}
