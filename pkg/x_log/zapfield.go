// file:mini/pkg/x_log/zapfield.go
package x_log

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Boolean
func FBool(key string, val bool) Field     { return zap.Bool(key, val) }
func FBools(key string, vals []bool) Field { return zap.Bools(key, vals) }

// Numeric
func FInt(key string, val int) Field     { return zap.Int(key, val) }
func FInts(key string, vals []int) Field { return zap.Ints(key, vals) }

func FInt32(key string, val int32) Field     { return zap.Int32(key, val) }
func FInt64(key string, val int64) Field     { return zap.Int64(key, val) }
func FInt64s(key string, vals []int64) Field { return zap.Int64s(key, vals) }

func FFloat32(key string, val float32) Field     { return zap.Float32(key, val) }
func FFloat32s(key string, vals []float32) Field { return zap.Float32s(key, vals) }
func FFloat64(key string, val float64) Field     { return zap.Float64(key, val) }
func FFloat64s(key string, vals []float64) Field { return zap.Float64s(key, vals) }

// String
func FString(key string, val string) Field         { return zap.String(key, val) }
func FStrings(key string, vals []string) Field     { return zap.Strings(key, vals) }
func FStringer(key string, val fmt.Stringer) Field { return zap.Stringer(key, val) }

// Time
func FTime(key string, val time.Time) Field             { return zap.Time(key, val) }
func FTimes(key string, vals []time.Time) Field         { return zap.Times(key, vals) }
func FDuration(key string, val time.Duration) Field     { return zap.Duration(key, val) }
func FDurations(key string, vals []time.Duration) Field { return zap.Durations(key, vals) }

// Binary / Byte
func FBinary(key string, val []byte) Field         { return zap.Binary(key, val) }
func FByteString(key string, val []byte) Field     { return zap.ByteString(key, val) }
func FByteStrings(key string, vals [][]byte) Field { return zap.ByteStrings(key, vals) }

// Error
func FError(err error) Field                  { return zap.Error(err) }
func FNamedError(key string, err error) Field { return zap.NamedError(key, err) }
func FErrors(key string, errs []error) Field  { return zap.Errors(key, errs) }

// Other / Special
func FAny(key string, val any) Field    { return zap.Any(key, val) }
func FObject(key string, val any) Field { return zap.Reflect(key, val) }
func FNamespace(key string) Field       { return zap.Namespace(key) }
func FStack(key string) Field           { return zap.Stack(key) }
func FSkip() Field                      { return zap.Skip() }
