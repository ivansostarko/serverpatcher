package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type FileLock struct {
	path string
	f    *os.File
}

func Acquire(path string) (*FileLock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("could not acquire lock %s (is another run active?): %w", path, err)
	}
	return &FileLock{path: path, f: f}, nil
}

func (l *FileLock) Release() error {
	if l == nil || l.f == nil {
		return nil
	}
	_ = syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
	return l.f.Close()
}
