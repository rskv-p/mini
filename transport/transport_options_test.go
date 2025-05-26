package transport

import (
	"os"
	"testing"
	"time"

	"github.com/rskv-p/mini/logger"
	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------
// Test Mocks
// ----------------------------------------------------

type testMetrics struct {
	counter string
	latency int64
}

func (m *testMetrics) IncCounter(name string) {
	m.counter = name
}
func (m *testMetrics) AddLatency(name string, ms int64) {
	m.latency = ms
}

type testLogger struct {
	fields map[string]any
}

func (l *testLogger) Debug(msg string, args ...any) {}
func (l *testLogger) Info(msg string, args ...any)  {}
func (l *testLogger) Warn(msg string, args ...any)  {}
func (l *testLogger) Error(msg string, args ...any) {}

func (l *testLogger) WithContext(contextID string) logger.ILogger {
	return l
}
func (l *testLogger) With(key string, value any) logger.LoggerEntry {
	if l.fields == nil {
		l.fields = make(map[string]any)
	}
	l.fields[key] = value
	return &testLoggerEntry{fields: l.fields}
}
func (l *testLogger) SetLevel(level string) {}
func (l *testLogger) Clone() logger.ILogger { return l }

type testLoggerEntry struct {
	fields map[string]any
}

func (e *testLoggerEntry) Debug(msg string, args ...any) {}
func (e *testLoggerEntry) Info(msg string, args ...any)  {}
func (e *testLoggerEntry) Warn(msg string, args ...any)  {}
func (e *testLoggerEntry) Error(msg string, args ...any) {}

func (e *testLoggerEntry) With(key string, value any) logger.LoggerEntry {
	if e.fields == nil {
		e.fields = make(map[string]any)
	}
	e.fields[key] = value
	return e
}
func (e *testLoggerEntry) Clone() logger.LoggerEntry {
	return e
}

var _ logger.ILogger = (*testLogger)(nil)
var _ logger.LoggerEntry = (*testLoggerEntry)(nil)

// ----------------------------------------------------
// Tests
// ----------------------------------------------------

func TestWithDefaults(t *testing.T) {
	opts := WithDefaults()
	assert.Equal(t, "default", opts.Subject)
	assert.False(t, opts.Debug)
	assert.False(t, opts.AutoReconnect)
	assert.NotNil(t, opts.RetryPolicies)
}

func TestOptionConstructors(t *testing.T) {
	var retried, failed, deadletterCalled bool

	handler := func(subject string, data []byte, err error) {
		deadletterCalled = true
	}
	retryFn := func(s string, i int, e error) { retried = true }
	failFn := func(s string, e error) { failed = true }

	metrics := &testMetrics{}
	log := &testLogger{}

	opts := WithDefaults()
	Subject("my.subject")(&opts)
	Addrs("127.0.0.1:4150")(&opts)
	Timeout(5 * time.Second)(&opts)
	WithDebug()(&opts)
	EnableReconnect()(&opts)
	WithMetrics(metrics)(&opts)
	WithLogger(log)(&opts)
	WithHooks(retryFn, failFn)(&opts)
	WithRetryPolicy("test.topic", RetryPolicy{MaxAttempts: 5, Delay: 100 * time.Millisecond})(&opts)
	WithRetry(3, 200*time.Millisecond)(&opts)
	WithDeadLetterHandler(handler)(&opts)

	assert.Equal(t, "my.subject", opts.Subject)
	assert.Equal(t, []string{"127.0.0.1:4150"}, opts.Addrs)
	assert.Equal(t, 5*time.Second, opts.Timeout)
	assert.True(t, opts.Debug)
	assert.True(t, opts.AutoReconnect)
	assert.Equal(t, metrics, opts.Metrics)
	assert.Equal(t, log, opts.Logger)
	assert.Contains(t, opts.RetryPolicies, "test.topic")
	assert.Contains(t, opts.RetryPolicies, "*")
	assert.NotNil(t, opts.OnRetry)
	assert.NotNil(t, opts.OnFailure)
	assert.NotNil(t, opts.DeadLetterHandler)

	opts.OnRetry("foo", 1, nil)
	opts.OnFailure("foo", nil)
	opts.DeadLetterHandler("foo", nil, nil)

	assert.True(t, retried)
	assert.True(t, failed)
	assert.True(t, deadletterCalled)
}

func TestFromEnv(t *testing.T) {
	os.Setenv("SRV_BUS_ADDR", "localhost:5000")
	os.Setenv("SRV_BUS_SUBJECT", "env.subject")
	os.Setenv("SRV_BUS_TIMEOUT", "1500")
	os.Setenv("SRV_BUS_DEBUG", "true")

	defer func() {
		os.Unsetenv("SRV_BUS_ADDR")
		os.Unsetenv("SRV_BUS_SUBJECT")
		os.Unsetenv("SRV_BUS_TIMEOUT")
		os.Unsetenv("SRV_BUS_DEBUG")
	}()

	opts := WithDefaults()
	FromEnv()(&opts)

	assert.Equal(t, []string{"localhost:5000"}, opts.Addrs)
	assert.Equal(t, "env.subject", opts.Subject)
	assert.Equal(t, 1500*time.Millisecond, opts.Timeout)
	assert.True(t, opts.Debug)
}

func TestWithEnvPrefix(t *testing.T) {
	os.Setenv("CUSTOM_ADDR", "1.2.3.4:9999")
	os.Setenv("CUSTOM_SUBJECT", "prefix.subj")
	os.Setenv("CUSTOM_TIMEOUT", "2000")
	os.Setenv("CUSTOM_DEBUG", "1")

	defer func() {
		os.Unsetenv("CUSTOM_ADDR")
		os.Unsetenv("CUSTOM_SUBJECT")
		os.Unsetenv("CUSTOM_TIMEOUT")
		os.Unsetenv("CUSTOM_DEBUG")
	}()

	opts := WithDefaults()
	WithEnvPrefix("CUSTOM_")(&opts)

	assert.Equal(t, []string{"1.2.3.4:9999"}, opts.Addrs)
	assert.Equal(t, "prefix.subj", opts.Subject)
	assert.Equal(t, 2*time.Second, opts.Timeout)
	assert.True(t, opts.Debug)
}
