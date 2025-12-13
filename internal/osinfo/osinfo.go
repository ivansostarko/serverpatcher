// Package osinfo provides detection of Linux distribution information
// by parsing /etc/os-release (and falls back to other methods if needed).
package osinfo

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Info struct {
	Name       string `json:"name"`
	ID         string `json:"id"`
	VersionID  string `json:"version_id"`
	PrettyName string `json:"pretty_name"`
	Like       string `json:"like"`
}

// unquote removes surrounding quotes from values in /etc/os-release
// Handles both "double quoted" and 'single quoted' values, as well as unquoted ones.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// Detect reads and parses /etc/os-release to determine the OS/distribution.
func Detect() (*Info, error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("open /etc/os-release: %w", err)
	}
	defer f.Close()

	m := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first '=' only
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := unquote(parts[1])
		m[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read /etc/os-release: %w", err)
	}

	info := &Info{
		Name:       m["NAME"],
		ID:         strings.ToLower(m["ID"]),
		VersionID:  m["VERSION_ID"],
		PrettyName: m["PRETTY_NAME"],
		Like:       strings.ToLower(m["ID_LIKE"]),
	}

	// Fallback: if NAME is empty, try PRETTY_NAME
	if info.Name == "" {
		info.Name = info.PrettyName
	}

	return info, nil
}

// IsLike checks if the OS is like a certain distribution family.
// Examples:
//   - Ubuntu is like "debian"
//   - Fedora is like "rhel fedora"
//   - Arch is like "archlinux arch"
func (i *Info) IsLike(token string) bool {
	token = strings.ToLower(token)

	if i.ID == token {
		return true
	}

	// Check ID_LIKE field (space-separated)
	for _, t := range strings.Fields(i.Like) {
		if t == token {
			return true
		}
	}

	return false
}

// Convenience methods
func (i *Info) IsDebianLike() bool   { return i.IsLike("debian") }
func (i *Info) IsUbuntu() bool      { return i.ID == "ubuntu" }
func (i *Info) IsRHELLike() bool    { return i.IsLike("rhel") || i.IsLike("fedora") }
func (i *Info) IsArchLinux() bool   { return i.ID == "arch" || i.IsLike("arch") }
func (i *Info) IsAlmaLinux() bool   { return i.ID == "almalinux" }
func (i *Info) IsRockyLinux() bool  { return i.ID == "rocky" }
func (i *Info) IsCentOS() bool      { return i.ID == "centos" }
func (i *Info) IsFedora() bool      { return i.ID == "fedora" }