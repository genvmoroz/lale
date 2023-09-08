package chatgpt

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewRepo(t *testing.T) {
	got, err := NewRepo("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")
	require.NoError(t, err)

	_, _ = got.GenerateSentences(context.Background(), "suspicion", 5)
}
