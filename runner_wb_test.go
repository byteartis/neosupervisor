//nolint:testpackage // Test it's easier with white-box testing
package neosupervisor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapRunner(t *testing.T) {
	t.Parallel()

	t.Run("when the name is not valid", func(t *testing.T) {
		t.Parallel()

		r := wrapRunner(context.Background(), new(MockRunner), WithRunnerName("invalid name"))

		assert.Empty(t, r.name)
	})

	t.Run("when the name is valid", func(t *testing.T) {
		t.Parallel()

		r := wrapRunner(context.Background(), new(MockRunner), WithRunnerName("valid_name"))

		assert.Equal(t, "valid_name", r.name)
	})
}
