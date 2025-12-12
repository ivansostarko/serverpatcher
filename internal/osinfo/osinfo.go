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

func Detect() (*Info, error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("open /etc/os-release: %w", err)
	}
	defer f.Close()

	m := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := parts[0]
		v := strings.Trim(parts[1], """)
		m[k] = v
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read /etc/os-release: %w", err)
	}

	return &Info{
		Name:       m["NAME"],
		ID:         m["ID"],
		VersionID:  m["VERSION_ID"],
		PrettyName: m["PRETTY_NAME"],
		Like:       m["ID_LIKE"],
	}, nil
}

func (i *Info) IsLike(token string) bool {
	token = strings.ToLower(token)
	if strings.ToLower(i.ID) == token {
		return true
	}
	for _, t := range strings.Fields(strings.ToLower(i.Like)) {
		if t == token {
			return true
		}
	}
	return false
}
