package highscore

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestDefaultPath checks platform selection without using real user data.
func TestDefaultPath(t *testing.T) {
	home := t.TempDir()
	data := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("LOCALAPPDATA", data)
	for _, goos := range []string{"linux", "windows", "darwin", "ios"} {
		for _, xdg := range []string{"", "relative/data", data} {
			t.Run(goos+"/"+xdg, func(t *testing.T) {
				t.Setenv("XDG_DATA_HOME", xdg)
				want := filepath.Join(home, ".local", "share", "geckoblu", "pacman", "highscore")
				switch goos {
				case "linux":
					if xdg == data {
						want = filepath.Join(data, "geckoblu", "pacman", "highscore")
					}
				case "windows":
					want = filepath.Join(data, "geckoblu", "pacman", "highscore")
				case "darwin", "ios":
					want = filepath.Join(home, "Library", "Application Support", "geckoblu", "pacman", "highscore")
				}
				got, err := defaultPath(goos)
				if err != nil || got != want {
					t.Fatalf("defaultPath(%q) = %q, %v; want %q", goos, got, err, want)
				}
				if goos == runtime.GOOS {
					got, err = DefaultPath()
					if err != nil || got != want {
						t.Fatalf("DefaultPath() = %q, %v; want %q", got, err, want)
					}
				}
			})
		}
	}
}

// TestDefaultPathRejectsUnavailableDirectories checks missing roots and mobile context.
func TestDefaultPathRejectsUnavailableDirectories(t *testing.T) {
	for _, value := range []string{"", "relative"} {
		t.Setenv("HOME", value)
		t.Setenv("LOCALAPPDATA", value)
		t.Setenv("XDG_DATA_HOME", "")
		for _, goos := range []string{"linux", "windows", "darwin", "ios", "android", "unsupported"} {
			if path, err := defaultPath(goos); err == nil || path != "" {
				t.Fatalf("expected no usable path for %s with root %q, got %q, %v", goos, value, path, err)
			}
		}
	}
}

// TestAndroidRequiresPrivatePath checks environment variables never select shared storage.
func TestAndroidRequiresPrivatePath(t *testing.T) {
	t.Setenv("HOME", "/sdcard")
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if path, err := defaultPath("android"); err == nil || path != "" {
		t.Fatalf("Android must require an explicit app-private path, got %q, %v", path, err)
	}
}

// TestStoreRoundTrip checks first-run behavior, restart loading, and unchanged-record writes.
func TestStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "geckoblu", "pacman", "highscore")
	store := New(path)
	if score, err := store.Load(); err != nil || score != 0 {
		t.Fatalf("missing file: got %d, %v", score, err)
	}
	if err := store.Save(12340); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "12340\n" {
		t.Fatalf("saved contents = %q, %v", data, err)
	}
	restarted := New(path)
	if score, err := restarted.Load(); err != nil || score != 12340 {
		t.Fatalf("restart: got %d, %v", score, err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, score := range []int{12340, 100, 0} {
		if err := restarted.Save(score); err != nil {
			t.Fatal(err)
		}
	}
	after, err := os.Stat(path)
	if err != nil || !os.SameFile(before, after) {
		t.Fatalf("unchanged or lower scores should not replace the file: %v", err)
	}
	if err := restarted.Save(15000); err != nil {
		t.Fatal(err)
	}
	if score, err := New(path).Load(); err != nil || score != 15000 {
		t.Fatalf("updated score: got %d, %v", score, err)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 || entries[0].Name() != "highscore" {
		t.Fatalf("unexpected files after save: %v, %v", entries, err)
	}
}

// TestLoadRejectsInvalidScores checks malformed, negative, and overflowing contents.
func TestLoadRejectsInvalidScores(t *testing.T) {
	path := filepath.Join(t.TempDir(), "highscore")
	for _, contents := range []string{"", "not a score", "-10\n", "999999999999999999999999999999"} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		store := New(path)
		if _, err := store.Load(); err == nil {
			t.Fatalf("expected error for %q", contents)
		}
		if err := store.Save(10); err != nil {
			t.Fatalf("could not replace invalid score: %v", err)
		}
	}
	if _, err := New(t.TempDir()).Load(); err == nil {
		t.Fatal("reading a directory should report an error")
	}
}

// TestSaveRetriesAfterFailure checks failed writes do not mark a score as saved.
func TestSaveRetriesAfterFailure(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(parent, []byte("block directory creation"), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "highscore")
	store := New(path)
	if err := store.Save(100); err == nil {
		t.Fatal("expected directory creation error")
	}
	if err := os.Remove(parent); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(100); err != nil {
		t.Fatal(err)
	}
	if score, err := New(path).Load(); err != nil || score != 100 {
		t.Fatalf("retry: got %d, %v", score, err)
	}
	if err := store.Save(-1); err == nil {
		t.Fatal("negative score should be rejected")
	}
}

// TestSaveCleansTemporaryFile checks a failed replacement leaves no temporary files.
func TestSaveCleansTemporaryFile(t *testing.T) {
	parent := t.TempDir()
	path := filepath.Join(parent, "highscore")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := New(path).Save(100); err == nil {
		t.Fatal("expected replacement of a directory to fail")
	}
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 1 || entries[0].Name() != "highscore" || !entries[0].IsDir() {
		t.Fatalf("failed save changed the destination or left temporary files: %v, %v", entries, err)
	}
}
