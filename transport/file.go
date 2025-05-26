// file: mini/transport/file.go
package transport

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/constant"
)

// ----------------------------------------------------
// File chunk metadata
// ----------------------------------------------------

type FileChunk struct {
	FileID     string
	Index      int
	Total      int
	IsLast     bool
	Filename   string
	Mime       string
	ChunkBytes []byte
}

// ----------------------------------------------------
// Sending file in chunks
// ----------------------------------------------------

func (t *Transport) SendFile(msg codec.IMessage, subject string, file []byte, chunkSize int) error {
	fileSize := len(file)
	total := chunkCount(fileSize, chunkSize)
	reader := bytes.NewReader(file)

	fileID := msg.GetContextID()
	if fileID == "" {
		fileID = generateFileID()
		msg.SetContextID(fileID)
	}

	filename := msg.GetString("filename")
	mime := msg.GetString("mime")

	for index := 0; index < total; index++ {
		size := chunkSize
		if rem := fileSize - index*size; rem < size {
			size = rem
		}
		chunk := make([]byte, size)
		if _, err := reader.Read(chunk); err != nil {
			return fmt.Errorf("read chunk %d: %w", index, err)
		}

		chunkMsg := codec.NewMessage(constant.MessageTypeStream)
		chunkMsg.SetContextID(fileID)
		chunkMsg.Set("fileID", fileID)
		chunkMsg.Set("chunkIndex", index)
		chunkMsg.Set("chunkTotal", total)
		chunkMsg.Set("fileSize", fileSize)
		chunkMsg.Set("isLast", index == total-1)
		chunkMsg.Set("fileChunk", chunk)
		if filename != "" {
			chunkMsg.Set("filename", filename)
		}
		if mime != "" {
			chunkMsg.Set("mime", mime)
		}

		data, err := codec.Marshal(chunkMsg)
		if err != nil {
			return fmt.Errorf("marshal chunk %d: %w", index, err)
		}

		if t.opts.Debug {
			fmt.Printf("[file] → %s | chunk %d/%d | size: %d | isLast: %v | fileID: %s\n",
				subject, index+1, total, len(chunk), index == total-1, fileID)
		}

		if err := t.Publish(subject, data); err != nil {
			return fmt.Errorf("publish chunk %d: %w", index, err)
		}
	}
	return nil
}

// ----------------------------------------------------
// Receive file helpers (chunk aggregation)
// ----------------------------------------------------

type FileReceiverHooks struct {
	OnChunk    func(FileChunk)
	OnComplete func([]byte, FileChunk)
	OnTimeout  func(fileID string)
}

func ReceiveFile(handler func([]byte, FileChunk, bool) error) MsgHandler {
	return ReceiveFileWithHooks(FileReceiverHooks{
		OnChunk: func(ch FileChunk) {
			_ = handler(ch.ChunkBytes, ch, false)
		},
		OnComplete: func(full []byte, meta FileChunk) {
			_ = handler(full, meta, true)
		},
	})
}

func ReceiveFileWithHooks(hooks FileReceiverHooks) MsgHandler {
	buffers := make(map[string][][]byte)
	meta := make(map[string]FileChunk)
	timers := make(map[string]*time.Timer)
	ttl := 60 * time.Second

	return func(data []byte) error {
		ch, err := decodeFileChunk(data)
		if err != nil {
			return err
		}
		fileID := ch.FileID

		meta[fileID] = ch
		buffers[fileID] = append(buffers[fileID], ch.ChunkBytes)

		if t := timers[fileID]; t != nil {
			t.Stop()
		}
		timers[fileID] = time.AfterFunc(ttl, func() {
			delete(buffers, fileID)
			delete(meta, fileID)
			delete(timers, fileID)
			if hooks.OnTimeout != nil {
				hooks.OnTimeout(fileID)
			}
		})

		if hooks.OnChunk != nil {
			hooks.OnChunk(ch)
		}
		if ch.IsLast && len(buffers[fileID]) == ch.Total {
			full := bytes.Join(buffers[fileID], nil)
			if hooks.OnComplete != nil {
				hooks.OnComplete(full, ch)
			}
			delete(buffers, fileID)
			delete(meta, fileID)
			if t := timers[fileID]; t != nil {
				t.Stop()
				delete(timers, fileID)
			}
		}
		return nil
	}
}

// FileChunkHandler handles full or partial file reception.
type FileChunkHandler func(context.Context, FileChunk, bool) error

func ReceiveFileHandler(handler FileChunkHandler) MsgHandler {
	buffers := make(map[string][][]byte)
	meta := make(map[string]FileChunk)
	timers := make(map[string]*time.Timer)
	ttl := 60 * time.Second

	return func(data []byte) error {
		ch, err := decodeFileChunk(data)
		if err != nil {
			return err
		}
		fileID := ch.FileID
		meta[fileID] = ch
		buffers[fileID] = append(buffers[fileID], ch.ChunkBytes)

		if t := timers[fileID]; t != nil {
			t.Stop()
		}
		timers[fileID] = time.AfterFunc(ttl, func() {
			delete(buffers, fileID)
			delete(meta, fileID)
			delete(timers, fileID)
			fmt.Printf("[receive] TTL expired: %s\n", fileID)
		})

		ctx := context.Background()

		if ch.IsLast && len(buffers[fileID]) == ch.Total {
			full := bytes.Join(buffers[fileID], nil)
			ch.ChunkBytes = full
			ch.IsLast = true
			delete(buffers, fileID)
			delete(meta, fileID)
			if t := timers[fileID]; t != nil {
				t.Stop()
				delete(timers, fileID)
			}
			return handler(ctx, ch, true)
		}

		return handler(ctx, ch, false)
	}
}

// ----------------------------------------------------
// Utility functions
// ----------------------------------------------------

func decodeFileChunk(data []byte) (FileChunk, error) {
	msg := codec.NewMessage("")
	if err := codec.Unmarshal(data, msg); err != nil {
		return FileChunk{}, err
	}

	raw, ok := msg.Get("fileChunk")
	if !ok {
		return FileChunk{}, fmt.Errorf("missing fileChunk")
	}

	var chunkBytes []byte
	switch v := raw.(type) {
	case []byte:
		chunkBytes = v
	case string: // ← base64 encoded []byte
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return FileChunk{}, fmt.Errorf("base64 decode failed: %w", err)
		}
		chunkBytes = decoded
	default:
		return FileChunk{}, fmt.Errorf("invalid fileChunk type")
	}

	return FileChunk{
		FileID:     msg.GetString("fileID"),
		Index:      int(msg.GetInt("chunkIndex")),
		Total:      int(msg.GetInt("chunkTotal")),
		IsLast:     msg.GetBool("isLast"),
		Filename:   msg.GetString("filename"),
		Mime:       msg.GetString("mime"),
		ChunkBytes: chunkBytes,
	}, nil
}
func chunkCount(size, chunkSize int) int {
	if size <= 0 || chunkSize <= 0 {
		return 0
	}
	return (size + chunkSize - 1) / chunkSize
}

func generateFileID() string {
	return fmt.Sprintf("file-%d", time.Now().UnixNano())
}
