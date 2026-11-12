// Package logging provides bounded, private operational log files.
package logging

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// Writer rotates before a write would exceed the configured byte limit.
// Backups are numbered .1 (newest) through .N. One process owns the file.
type Writer struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	backups  int
	file     *os.File
	size     int64
}

func Open(path string, maxBytes int64, backups int) (*Writer, error) {
	if maxBytes <= 0 || backups < 1 || backups > 100 {
		return nil, errors.New("log size must be positive and backups must be between 1 and 100")
	}
	w := &Writer{path: path, maxBytes: maxBytes, backups: backups}
	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Writer) open() error {
	if info, err := os.Lstat(w.path); err == nil {
		if !info.Mode().IsRegular() {
			return errors.New("log destination must be a regular file")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}
	w.file, w.size = file, info.Size()
	return nil
}

func (w *Writer) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return 0, os.ErrClosed
	}
	if int64(len(data)) > w.maxBytes {
		return 0, fmt.Errorf("log record exceeds %d bytes", w.maxBytes)
	}
	if w.size+int64(len(data)) > w.maxBytes {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(data)
	w.size += int64(n)
	if n != len(data) && err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

func (w *Writer) rotate() error {
	if err := w.file.Close(); err != nil {
		return err
	}
	w.file = nil
	for i := w.backups; i >= 1; i-- {
		old := w.path
		if i > 1 {
			old = fmt.Sprintf("%s.%d", w.path, i-1)
		}
		if err := os.Rename(old, fmt.Sprintf("%s.%d", w.path, i)); err != nil && !errors.Is(err, os.ErrNotExist) {
			// Keep logging usable if a backup cannot be moved.
			_ = w.open()
			return err
		}
	}
	return w.open()
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}
