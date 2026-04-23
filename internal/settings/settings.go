package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Settings is the user-configurable state persisted across runs. Keep this
// struct JSON-stable — fields are read verbatim from disk.
type Settings struct {
	FontSize int `json:"fontSize"`
}

// Defaults returns the baseline values used on a fresh install, when the
// config file is missing, or when a field is out of range.
func Defaults() Settings {
	return Settings{FontSize: 12}
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "pik", "settings.json"), nil
}

// Load reads the config file. Missing file → defaults; malformed file is
// reported but defaults are still returned so the app can start.
func Load() (Settings, error) {
	s := Defaults()
	p, err := configPath()
	if err != nil {
		return s, nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, nil
		}
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return Defaults(), err
	}
	return Sanitize(s), nil
}

// Save writes the settings atomically (temp file + rename) to avoid leaving
// a half-written JSON if the process is killed mid-write.
func Save(s Settings) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(Sanitize(s), "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// Sanitize clamps out-of-range values to sane ones. Called on both read and
// write so a hand-edited config can't break the UI.
func Sanitize(s Settings) Settings {
	if s.FontSize < 8 {
		s.FontSize = 8
	}
	if s.FontSize > 24 {
		s.FontSize = 24
	}
	return s
}
