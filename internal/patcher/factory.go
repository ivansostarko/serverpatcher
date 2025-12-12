package patcher

import (
	"fmt"

	"github.com/serverpatcher/serverpatcher/internal/executil"
	"github.com/serverpatcher/serverpatcher/internal/osinfo"
)

func Select(info *osinfo.Info) (Patcher, error) {
	if _, ok := executil.LookPathAny("apt-get"); ok {
		return &Apt{}, nil
	}
	if _, ok := executil.LookPathAny("dnf"); ok {
		return &Dnf{}, nil
	}
	if _, ok := executil.LookPathAny("yum"); ok {
		return &Yum{}, nil
	}
	if _, ok := executil.LookPathAny("zypper"); ok {
		return &Zypper{}, nil
	}
	if _, ok := executil.LookPathAny("pacman"); ok {
		return &Pacman{}, nil
	}
	if _, ok := executil.LookPathAny("apk"); ok {
		return &Apk{}, nil
	}
	return nil, fmt.Errorf("no supported package manager detected for %s", info.PrettyName)
}
