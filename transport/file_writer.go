// file: mini/transport/file_writer.go
package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// IStreamFileWriter defines a chunk-based file writer.
type IStreamFileWriter interface {
	WriteChunk(chunk FileChunk) error
	Close() error
}

// Ensure interface compliance
var (
	_ IStreamFileWriter = (*MemoryFileWriter)(nil)
	_ IStreamFileWriter = (*DiskFileWriter)(nil)
)

//
// ─────────────────────────────────────────────────────────────
// In-memory implementation
// ─────────────────────────────────────────────────────────────
//

type MemoryFileWriter struct {
	buf    *bytes.Buffer
	closed bool
	mu     sync.Mutex
}

// NewMemoryFileWriter returns a new memory buffer-based writer.
func NewMemoryFileWriter() *MemoryFileWriter {
	return &MemoryFileWriter{
		buf: new(bytes.Buffer),
	}
}

func (w *MemoryFileWriter) WriteChunk(chunk FileChunk) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return io.ErrClosedPipe
	}
	_, err := w.buf.Write(chunk.ChunkBytes)
	return err
}

func (w *MemoryFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	return nil
}

// Bytes returns the full buffer (read-only).
func (w *MemoryFileWriter) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Bytes()
}

//
// ─────────────────────────────────────────────────────────────
// Disk-based implementation
// ─────────────────────────────────────────────────────────────
//

type DiskFileWriter struct {
	file   *os.File
	path   string
	closed bool
	mu     sync.Mutex
}

// NewDiskFileWriter creates a file writer for a given path.
func NewDiskFileWriter(path string) (*DiskFileWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &DiskFileWriter{file: f, path: path}, nil
}

func (w *DiskFileWriter) WriteChunk(chunk FileChunk) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return io.ErrClosedPipe
	}
	_, err := w.file.Write(chunk.ChunkBytes)
	return err
}

func (w *DiskFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	return w.file.Close()
}

// Path returns the file path used for writing (optional).
func (w *DiskFileWriter) Path() string {
	return w.path
}

//
// ─────────────────────────────────────────────────────────────
// ReceiveFileRouter utility
// ─────────────────────────────────────────────────────────────
//
// ReceiveFileRouter creates a MsgHandler that assembles chunks into a writer.
// newWriter is called once per fileID.
//

func ReceiveFileRouter(newWriter func(meta FileChunk) (IStreamFileWriter, error)) MsgHandler {
	active := make(map[string]IStreamFileWriter)
	var mu sync.Mutex

	return ReceiveFileHandler(func(ctx context.Context, chunk FileChunk, isLast bool) error {
		mu.Lock()
		writer, ok := active[chunk.FileID]
		if !ok {
			var err error
			writer, err = newWriter(chunk)
			if err != nil {
				mu.Unlock()
				return fmt.Errorf("create writer for %s: %w", chunk.FileID, err)
			}
			active[chunk.FileID] = writer
		}
		mu.Unlock()

		if err := writer.WriteChunk(chunk); err != nil {
			return fmt.Errorf("write chunk %d: %w", chunk.Index, err)
		}

		if isLast {
			mu.Lock()
			delete(active, chunk.FileID)
			mu.Unlock()
			if err := writer.Close(); err != nil {
				return fmt.Errorf("close writer %s: %w", chunk.FileID, err)
			}
		}
		return nil
	})
}
