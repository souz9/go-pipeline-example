package pair_with_error

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPairWithError(t *testing.T) {
	result, err := Run(5, -1)
	assert.NoError(t, err)

	assert.Len(t, result, 5)
	assert.Contains(t, result, "0")
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "2")
	assert.Contains(t, result, "3")
	assert.Contains(t, result, "4")

	t.Run("simulate error", func(t *testing.T) {
		result, err := Run(5, 1)
		assert.Error(t, err)

		assert.Len(t, result, 2)
		assert.Contains(t, result, "0")
		assert.Contains(t, result, "1")
	})
}
