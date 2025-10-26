package pdf

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteValueTo(t *testing.T) {
	cases := []struct {
		in       any
		expected string
	}{
		{nil, "null"},
		{true, "true"},

		{1, "1"},
		{2.1, "2.1"},

		{"abc", "(abc)"},
		{"a(b)c", `(a\(b\)c)`},
		{"å", "(å)"},

		{Name("abc"), "/abc"},
		{Name("a c"), "/a#20c"},
		{Name("a#c"), "/a#23c"},

		{[]any{}, "[]"},
		{[]any{1, 2, false, "hello"}, "[1 2 false (hello)]"},

		{map[string]any{}, "<<>>"},
		{map[string]any{"Length": 100}, "<</Length 100>>"},

		{[]byte{0x00, 0x01, 0x02}, "<000102>"},

		{objectReference{}, "0 0 R"},
		{objectReference{nil, 2, 3}, "2 3 R"},

		{NewStream([]byte("abc")), "<</Length 3>>\nstream\nabc\nendstream\n"},
	}

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			var s strings.Builder
			n, err := WriteValueTo(&s, c.in)

			assert.NoError(t, err)
			assert.Equal(t, int64(len(c.expected)), n)
			assert.Equal(t, c.expected, s.String())
		})
	}
}

func TestWriteObjectStartEnd(t *testing.T) {
	var s strings.Builder
	obj := objectID{nil, 4, 5}
	n, err := obj.writeStartTo(&s)
	expected := "\n4 5 obj\n"
	assert.NoError(t, err)
	assert.Equal(t, int64(len(expected)), n)
	assert.Equal(t, expected, s.String())

	s.Reset()
	n, err = obj.writeEndTo(&s)
	expected = "\nendobj\n"
	assert.NoError(t, err)
	assert.Equal(t, int64(len(expected)), n)
	assert.Equal(t, expected, s.String())
}

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
