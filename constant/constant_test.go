package constant_test

import (
	"testing"

	"github.com/rskv-p/mini/constant"
	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	errs := []error{
		constant.ErrBadRequest,
		constant.ErrNotFound,
		constant.ErrEmptyMessage,
		constant.ErrMissingHandler,
		constant.ErrNoRegistry,
		constant.ErrNoAvailableNodes,
		constant.ErrEmptyNodeList,
		constant.ErrInvalidPath,
	}
	for _, err := range errs {
		assert.Error(t, err)
		assert.NotEmpty(t, err.Error())
	}
}

func TestConstants_Values(t *testing.T) {
	assert.Equal(t, "bus://127.0.0.1:4222", constant.DefaultURL)
	assert.Equal(t, 2*1024*1024, constant.MaxFileChunkSize)
	assert.Equal(t, "request", constant.MessageTypeRequest)
	assert.Equal(t, 0, constant.HealthOK)
	assert.Equal(t, 1, constant.StatusWarning)
	assert.Equal(t, 504, constant.StatusTimeout)
	assert.Equal(t, "memory_critical", constant.MemoryCriticalKey)
	assert.Equal(t, "run", constant.OrganizationName)
}
