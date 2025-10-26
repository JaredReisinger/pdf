package pdf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteStuffTo(t *testing.T) {
	var s strings.Builder

	n, err := writeStuffTo(&s,
		[]byte("abc"),
		[]byte("123"),
		[]byte("abc"),
	)

	assert.NoError(t, err)
	assert.Equal(t, int64(9), n)
	assert.Equal(t, "abc123abc", s.String())
}
