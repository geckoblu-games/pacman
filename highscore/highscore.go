// Package highscore persists the high score independently of gameplay and rendering.
package highscore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Store reads and replaces a high-score file using a synced temporary file.
// It is intended for use by a single application instance.
type Store struct {
	// path is the destination file.
	path string
	// saved is the last score successfully loaded or saved.
	saved int
}

// DefaultPath resolves the platform's persistent user-data path.
// Android hosts must instead pass an app-private file path to New.
func DefaultPath() (string, error) {
	return defaultPath(runtime.GOOS)
}

// defaultPath selects the data directory without guessing mobile sandbox locations.
func defaultPath(goos string) (string, error) {
	var dir string
	switch goos {
	case "linux":
		dir = os.Getenv("XDG_DATA_HOME")
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(os.Getenv("HOME"), ".local", "share")
		}
	case "windows":
		dir = os.Getenv("LOCALAPPDATA")
	case "darwin", "ios":
		// UserHomeDir can fall back to "/" on iOS, outside the app's sandbox.
		home := os.Getenv("HOME")
		if !filepath.IsAbs(home) {
			return "", fmt.Errorf("resolve high-score directory on %s: HOME must be absolute", goos)
		}
		dir = filepath.Join(home, "Library", "Application Support")
	case "android":
		return "", errors.New("Android requires an app-private file path: pass a file inside Context.getFilesDir() to highscore.New")
	default:
		return "", fmt.Errorf("no default high-score directory on %s: pass an explicit file path to highscore.New", goos)
	}
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("resolve high-score directory on %s: data directory must be absolute", goos)
	}
	return filepath.Join(dir, "geckoblu", "pacman", "highscore"), nil
}

// New creates a store for path without accessing the filesystem.
// Mobile hosts may supply a file in their native app-private persistent directory.
func New(path string) *Store {
	return &Store{path: path}
}

// Load returns the stored nonnegative score, or zero if the file does not exist.
// Other read errors and malformed contents are returned to the caller.
func (store *Store) Load() (int, error) {
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read high score: %w", err)
	}
	score, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse high score in %q: %w", store.path, err)
	}
	if score < 0 {
		return 0, fmt.Errorf("negative high score in %q", store.path)
	}
	store.saved = score
	return score, nil
}

// Save writes only scores higher than the last successful load or save.
// The temporary file is synced before replacement in the same directory.
// On Unix the replacement is atomic and the containing directory is also synced.
func (store *Store) Save(score int) (err error) {
	if score < 0 {
		return fmt.Errorf("cannot save negative high score: %d", score)
	}
	if score <= store.saved {
		return nil
	}
	dir := filepath.Dir(store.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create high-score directory: %w", err)
	}
	file, err := os.CreateTemp(dir, ".highscore-*")
	if err != nil {
		return fmt.Errorf("create temporary high-score file: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("remove temporary high-score file: %w", removeErr))
		}
	}()
	_, writeErr := file.WriteString(strconv.Itoa(score) + "\n")
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return fmt.Errorf("write high score: %w", err)
	}
	if err := os.Rename(file.Name(), store.path); err != nil {
		return fmt.Errorf("replace high score: %w", err)
	}
	// Windows cannot sync directory handles opened by os.Open.
	if runtime.GOOS != "windows" {
		directory, err := os.Open(dir)
		if err != nil {
			return fmt.Errorf("open high-score directory for sync: %w", err)
		}
		syncErr := directory.Sync()
		closeErr = directory.Close()
		if err := errors.Join(syncErr, closeErr); err != nil {
			return fmt.Errorf("sync high-score directory: %w", err)
		}
	}
	store.saved = score
	return nil
}
