// file: mini/transport/file_test.go
package transport

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rskv-p/mini/codec"
	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------
// Mocks
// ----------------------------------------------------

type mockTransport struct {
	*Transport
	published [][]byte
	errOn     int
}

func newMockTransport() *mockTransport {
	tr := &Transport{}
	return &mockTransport{
		Transport: tr,
		published: make([][]byte, 0),
		errOn:     -1,
	}
}

func (m *mockTransport) overridePublish() {
	m.Transport.conn = &mockConnFile{
		published: &m.published,
		errOn:     m.errOn,
	}
}

type mockConnFile struct {
	published *[]byteSlice
	errOn     int
	count     int
}

type byteSlice = []byte

func (m *mockConnFile) Publish(subject string, data []byte) error {
	if m.errOn >= 0 && m.count == m.errOn {
		return fmt.Errorf("forced publish error")
	}
	*m.published = append(*m.published, data)
	m.count++
	return nil
}

func (m *mockConnFile) Request(string, []byte, time.Duration) (codec.IMessage, error) {
	return nil, nil
}
func (m *mockConnFile) Subscribe(string, MsgHandler) (*Subscription, error) {
	return nil, nil
}
func (m *mockConnFile) SubscribeOnce(string, MsgHandler, time.Duration) (*Subscription, error) {
	return nil, nil
}
func (m *mockConnFile) Close()            {}
func (m *mockConnFile) IsConnected() bool { return true }
func (m *mockConnFile) Ping() error       { return nil }

// ----------------------------------------------------
// Helpers
// ----------------------------------------------------

func createTestChunkMessage(fc FileChunk) []byte {
	msg := codec.NewMessage("")
	msg.Set("fileID", fc.FileID)
	msg.Set("chunkIndex", fc.Index)
	msg.Set("chunkTotal", fc.Total)
	msg.Set("isLast", fc.IsLast)
	msg.Set("fileChunk", fc.ChunkBytes)
	if fc.Filename != "" {
		msg.Set("filename", fc.Filename)
	}
	if fc.Mime != "" {
		msg.Set("mime", fc.Mime)
	}
	data, _ := codec.Marshal(msg)
	return data
}

// ----------------------------------------------------
// SendFile tests
// ----------------------------------------------------

func TestSendFile(t *testing.T) {
	mt := newMockTransport()
	mt.overridePublish()

	tr := mt.Transport
	msg := codec.NewMessage("")
	msg.Set("filename", "test.txt")
	msg.Set("mime", "text/plain")

	data := []byte("HelloWorld1234567890") // 20 bytes

	err := tr.SendFile(msg, "topic.test", data, 5)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(mt.published)) // 4 chunks expected
}

func TestSendFile_PublishError(t *testing.T) {
	mt := newMockTransport()
	mt.errOn = 2 // fail on 3rd chunk
	mt.overridePublish()

	tr := mt.Transport
	msg := codec.NewMessage("")
	data := []byte("chunkeddatahere") // 15 bytes → 4 chunks @ size=4

	err := tr.SendFile(msg, "fail.topic", data, 4)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish chunk")
}

// ----------------------------------------------------
// ReceiveFile tests
// ----------------------------------------------------

var timedOut string

func TestReceiveFileWithHooks(t *testing.T) {
	var received FileChunk
	var completed []byte

	hooks := FileReceiverHooks{
		OnChunk: func(fc FileChunk) { received = fc },
		OnComplete: func(full []byte, meta FileChunk) {
			completed = full
		},
		OnTimeout: func(fid string) { timedOut = fid },
	}

	handler := ReceiveFileWithHooks(hooks)
	data := createTestChunkMessage(FileChunk{
		FileID:     "abc123",
		Index:      0,
		Total:      1,
		IsLast:     true,
		ChunkBytes: []byte("DATA"),
	})

	err := handler(data)
	assert.NoError(t, err)
	assert.Equal(t, "DATA", string(received.ChunkBytes))
	assert.Equal(t, "DATA", string(completed))
}

func TestReceiveFileHandler(t *testing.T) {
	var result FileChunk
	var isFinal bool

	handler := ReceiveFileHandler(func(_ context.Context, fc FileChunk, final bool) error {
		result = fc
		isFinal = final
		return nil
	})

	data := createTestChunkMessage(FileChunk{
		FileID:     "file1",
		Index:      0,
		Total:      1,
		IsLast:     true,
		ChunkBytes: []byte("hello"),
	})

	err := handler(data)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(result.ChunkBytes))
	assert.True(t, isFinal)
}

// ----------------------------------------------------
// Decode and helper tests
// ----------------------------------------------------

func TestDecodeFileChunk_Errors(t *testing.T) {
	_, err := decodeFileChunk([]byte("not json"))
	assert.Error(t, err)

	msg := codec.NewMessage("")
	msg.Set("fileID", "x")
	msg.Set("chunkIndex", 0)
	msg.Set("chunkTotal", 1)
	msg.Set("isLast", true)
	msg.Set("fileChunk", "not bytes")

	data, _ := codec.Marshal(msg)
	_, err = decodeFileChunk(data)
	assert.Error(t, err)
}

func TestChunkCount(t *testing.T) {
	assert.Equal(t, 0, chunkCount(-1, 10))
	assert.Equal(t, 0, chunkCount(10, 0))
	assert.Equal(t, 3, chunkCount(25, 10))
}

func TestGenerateFileID(t *testing.T) {
	id := generateFileID()
	assert.Contains(t, id, "file-")
}
