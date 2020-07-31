package using_errgroup

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsingErrGroup(t *testing.T) {
	result, err := Run(context.Background(), 5, math.MaxInt32)
	assert.NoError(t, err)

	assert.Len(t, result, 5)
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "2")
	assert.Contains(t, result, "3")
	assert.Contains(t, result, "4")
	assert.Contains(t, result, "5")

	t.Run("simulate error", func(t *testing.T) {
		_, err := Run(context.Background(), 3, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Some error")
	})
}
