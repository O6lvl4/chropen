// Package profile resolves Google Chrome profile directories from
// email addresses and launches Chrome with --profile-directory pointed
// at the right one.
package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Info describes one Chrome profile directory.
type Info struct {
	Dir         string // "Default", "Profile 24", ...
	Email       string // first signed-in account, when known
	DisplayName string // user-set profile label
}

// BaseDir returns the Chrome user-data directory for the current OS.
// Override with CHROME_USER_DATA_DIR.
func BaseDir() (string, error) {
	if v := os.Getenv("CHROME_USER_DATA_DIR"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Google", "Chrome"), nil
	case "linux":
		// Try canonical paths in order.
		for _, candidate := range []string{
			filepath.Join(home, ".config", "google-chrome"),
			filepath.Join(home, ".config", "chromium"),
		} {
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
		return filepath.Join(home, ".config", "google-chrome"), nil
	case "windows":
		return filepath.Join(home, "AppData", "Local", "Google", "Chrome", "User Data"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// Binary returns the Chrome executable path. Override with CHROME_BIN.
func Binary() string {
	if v := os.Getenv("CHROME_BIN"); v != "" {
		return v
	}
	switch runtime.GOOS {
	case "darwin":
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	case "linux":
		return "google-chrome"
	case "windows":
		return "chrome.exe"
	}
	return "chrome"
}

// Resolve maps either an email or a literal directory name to a Chrome
// profile directory. Inputs without "@" are passed through as-is, so
// callers can pass either form transparently.
func Resolve(input string) (string, error) {
	if input == "" {
		return "", errors.New("empty profile identifier")
	}
	if !strings.Contains(input, "@") {
		return input, nil
	}
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	if dir, _ := lookupViaLocalState(base, input); dir != "" {
		return dir, nil
	}
	if dir, _ := lookupViaPreferences(base, input); dir != "" {
		return dir, nil
	}
	return "", fmt.Errorf("no Chrome profile found for email %q under %s", input, base)
}

func lookupViaLocalState(base, email string) (string, error) {
	b, err := os.ReadFile(filepath.Join(base, "Local State"))
	if err != nil {
		return "", err
	}
	var ls struct {
		Profile struct {
			InfoCache map[string]struct {
				UserName string `json:"user_name"`
			} `json:"info_cache"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(b, &ls); err != nil {
		return "", err
	}
	for dir, info := range ls.Profile.InfoCache {
		if strings.EqualFold(info.UserName, email) {
			return dir, nil
		}
	}
	return "", nil
}

func lookupViaPreferences(base, email string) (string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(base, e.Name(), "Preferences"))
		if err != nil {
			continue
		}
		var pref struct {
			AccountInfo []struct {
				Email string `json:"email"`
			} `json:"account_info"`
		}
		if err := json.Unmarshal(b, &pref); err != nil {
			continue
		}
		for _, ai := range pref.AccountInfo {
			if strings.EqualFold(ai.Email, email) {
				return e.Name(), nil
			}
		}
	}
	return "", nil
}

// List returns every Chrome profile detected under BaseDir().
func List() ([]Info, error) {
	base, err := BaseDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}
	var out []Info
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		switch name {
		case "System Profile", "Crashpad", "ShaderCache", "GraphiteDawnCache",
			"GrShaderCache", "Subresource Filter", "Safe Browsing", "OptimizationGuide":
			continue
		}
		b, err := os.ReadFile(filepath.Join(base, name, "Preferences"))
		if err != nil {
			continue
		}
		var pref struct {
			AccountInfo []struct {
				Email string `json:"email"`
			} `json:"account_info"`
			Profile struct {
				Name string `json:"name"`
			} `json:"profile"`
		}
		if err := json.Unmarshal(b, &pref); err != nil {
			continue
		}
		info := Info{Dir: name, DisplayName: pref.Profile.Name}
		if len(pref.AccountInfo) > 0 {
			info.Email = pref.AccountInfo[0].Email
		}
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Dir < out[j].Dir })
	return out, nil
}

// Open launches Chrome with --profile-directory and the given URLs.
// It returns immediately (does not wait for the browser to exit).
func Open(dir string, urls ...string) error {
	if dir == "" {
		return errors.New("empty profile dir")
	}
	args := []string{"--profile-directory=" + dir}
	args = append(args, urls...)
	cmd := exec.Command(Binary(), args...)
	return cmd.Start()
}

// OpenAs is a convenience: Resolve+Open in one call.
func OpenAs(emailOrDir string, urls ...string) error {
	dir, err := Resolve(emailOrDir)
	if err != nil {
		return err
	}
	return Open(dir, urls...)
}
