// file: mini/transport/file.go
package transport

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/constant"
)

// ----------------------------------------------------
// File chunk metadata
// ----------------------------------------------------

// FileChunk holds metadata and payload for a file chunk.
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

// SendFile splits a file into chunks and publishes each chunk.
func (t *Transport) SendFile(msg codec.IMessage, subject string, file []byte) error {
	fileSize := len(file)
	total := chunkCount(fileSize, constant.MaxFileChunkSize)
	reader := bytes.NewReader(file)

	fileID := msg.GetContextID()
	if fileID == "" {
		fileID = generateFileID()
		msg.SetContextID(fileID)
	}

	filename := msg.GetString("filename")
	mime := msg.GetString("mime")

	for index := 0; index < total; index++ {
		size := constant.MaxFileChunkSize
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
// Receiver helpers
// ----------------------------------------------------

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

type FileReceiverHooks struct {
	OnChunk    func(FileChunk)
	OnComplete func([]byte, FileChunk)
	OnTimeout  func(string)
}

func ReceiveFileWithHooks(hooks FileReceiverHooks) MsgHandler {
	buffers := make(map[string][][]byte)
	meta := make(map[string]FileChunk)
	timers := make(map[string]*time.Timer)
	ttl := 60 * time.Second

	return func(data []byte) error {
		msg := codec.NewMessage("")
		if err := codec.Unmarshal(data, msg); err != nil {
			return err
		}

		fileID := msg.GetString("fileID")
		index := int(msg.GetInt("chunkIndex"))
		total := int(msg.GetInt("chunkTotal"))
		isLast := msg.GetBool("isLast")

		raw, ok := msg.Get("fileChunk")
		if !ok {
			return fmt.Errorf("missing fileChunk")
		}
		chunkBytes, ok := raw.([]byte)
		if !ok {
			return fmt.Errorf("invalid fileChunk type")
		}

		ch := FileChunk{
			FileID:     fileID,
			Index:      index,
			Total:      total,
			IsLast:     isLast,
			Filename:   msg.GetString("filename"),
			Mime:       msg.GetString("mime"),
			ChunkBytes: chunkBytes,
		}
		meta[fileID] = ch
		buffers[fileID] = append(buffers[fileID], chunkBytes)

		if timer, exists := timers[fileID]; exists {
			timer.Stop()
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
		if isLast && len(buffers[fileID]) == total {
			full := bytes.Join(buffers[fileID], nil)
			defer func() {
				delete(buffers, fileID)
				delete(meta, fileID)
				if tmr, exists := timers[fileID]; exists {
					tmr.Stop()
					delete(timers, fileID)
				}
			}()
			if hooks.OnComplete != nil {
				hooks.OnComplete(full, ch)
			}
		}
		return nil
	}
}

// chunkCount returns number of chunks needed.
func chunkCount(size, chunkSize int) int {
	if size <= 0 || chunkSize <= 0 {
		return 0
	}
	return (size + chunkSize - 1) / chunkSize
}

// generateFileID produces a unique file identifier.
func generateFileID() string {
	return fmt.Sprintf("file-%d", time.Now().UnixNano())
}

// FileChunkHandler handles assembled file chunks.
type FileChunkHandler func(context.Context, FileChunk, bool) error

// ReceiveFileHandler wraps a FileChunkHandler into MsgHandler.
func ReceiveFileHandler(handler FileChunkHandler) MsgHandler {
	buffers := make(map[string][][]byte)
	meta := make(map[string]FileChunk)
	timers := make(map[string]*time.Timer)
	ttl := 60 * time.Second

	return func(data []byte) error {
		msg := codec.NewMessage("")
		if err := codec.Unmarshal(data, msg); err != nil {
			return err
		}

		fileID := msg.GetString("fileID")
		index := int(msg.GetInt("chunkIndex"))
		total := int(msg.GetInt("chunkTotal"))
		isLast := msg.GetBool("isLast")

		raw, ok := msg.Get("fileChunk")
		if !ok {
			return fmt.Errorf("missing fileChunk")
		}
		chunkBytes, ok := raw.([]byte)
		if !ok {
			return fmt.Errorf("invalid chunk type")
		}

		ch := FileChunk{
			FileID:     fileID,
			Index:      index,
			Total:      total,
			IsLast:     isLast,
			Filename:   msg.GetString("filename"),
			Mime:       msg.GetString("mime"),
			ChunkBytes: chunkBytes,
		}
		meta[fileID] = ch
		buffers[fileID] = append(buffers[fileID], chunkBytes)

		if tmr, exists := timers[fileID]; exists {
			tmr.Stop()
		}
		timers[fileID] = time.AfterFunc(ttl, func() {
			delete(buffers, fileID)
			delete(meta, fileID)
			delete(timers, fileID)
			fmt.Printf("[receive] TTL expired: %s\n", fileID)
		})

		ctx := context.Background()
		if isLast && len(buffers[fileID]) == total {
			full := bytes.Join(buffers[fileID], nil)
			ch.ChunkBytes = full
			ch.IsLast = true
			delete(buffers, fileID)
			delete(meta, fileID)
			if tmr, exists := timers[fileID]; exists {
				tmr.Stop()
				delete(timers, fileID)
			}
			return handler(ctx, ch, true)
		}
		return handler(ctx, ch, false)
	}
}
