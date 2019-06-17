package rate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSlideWindowStorageFromDSN_Redis(t *testing.T) {
	s, err := NewSlideWindowStorageFromDSN("redis://localhost:6379/0")
	assert.NoError(t, err)
	assert.IsType(t, &redisSlideWindowStorage{}, s)
}

func TestNewSlideWindowStorageFromDSN_InMemory(t *testing.T) {
	s, err := NewSlideWindowStorageFromDSN("inmemory://")
	assert.NoError(t, err)
	assert.IsType(t, &inMemorySlideWindowStorage{}, s)
}
